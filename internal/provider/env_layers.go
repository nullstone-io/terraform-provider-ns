package provider

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nullstone-io/module/platformdata"
)

// Shared implementation for the layered env data sources `ns_env_layout` and `ns_env_values`.
// See specs/env-layers.md.

// capabilityEnvEntryType is the element type of `capability_env` and `capability_secrets`.
var capabilityEnvEntryType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"capability": tftypes.String,
	"name":       tftypes.String,
	"value":      tftypes.String,
}}

// capabilityKeyEntryType is the element type of `capability_secret_keys` (no value).
var capabilityKeyEntryType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"capability": tftypes.String,
	"name":       tftypes.String,
}}

const invalidEnvKeyDetail = "An environment variable key can only contain letters, numbers, and the underscore character. It also can not begin with a number."

// capabilityEntry is one env var or secret emitted by a capability module.
type capabilityEntry struct {
	Capability string
	Name       string
	Value      string
}

// layeredEnvInput is the fully-known configuration of a layered env data source.
// `ns_env_layout` fills CapabilitySecrets/UserSecrets with empty values (keys only).
type layeredEnvInput struct {
	Platform           string
	Standard           map[string]string
	Cloud              map[string]string
	Otel               map[string]string
	CapabilityEnv      []capabilityEntry
	CapabilitySecrets  []capabilityEntry
	CapabilityPrefixes map[string]string
	UserEnv            map[string]string
	UserSecrets        map[string]string
}

// layeredEnvKey records where a key came from and the raw (pre-interpolation) value that won.
type layeredEnvKey struct {
	Template   string
	Source     string
	Capability string
	// SecretInput is true when the winning layer was a secrets input; its template is the secret value itself.
	SecretInput bool
}

// layeredEnvResult is the outcome of resolving the layers.
type layeredEnvResult struct {
	Platform platformdata.Platform
	Vars     EnvVars
	// Keys holds source/capability/template for every key in Vars.
	Keys map[string]layeredEnvKey
	// Classification (disjoint sets covering every key in Vars).
	Plain     []string
	Managed   []string
	Unmanaged []string
	K8s       []string
}

func (r layeredEnvResult) AllSecretKeys() []string {
	all := append(append([]string{}, r.Managed...), r.Unmanaged...)
	sort.Strings(all)
	return all
}

func (r layeredEnvResult) Sources() map[string]string {
	result := map[string]string{}
	for k, v := range r.Keys {
		result[k] = v.Source
	}
	return result
}

func (r layeredEnvResult) Capabilities() map[string]string {
	result := map[string]string{}
	for k, v := range r.Keys {
		if v.Source == platformdata.SourceCapability {
			result[k] = v.Capability
		}
	}
	return result
}

// PlainValues returns the post-interpolation values of plain keys.
func (r layeredEnvResult) PlainValues() map[string]string {
	result := map[string]string{}
	for _, k := range r.Plain {
		result[k] = r.Vars[k].Value
	}
	return result
}

// ManagedValues returns the post-interpolation values of managed secrets.
func (r layeredEnvResult) ManagedValues() map[string]string {
	result := map[string]string{}
	for _, k := range r.Managed {
		result[k] = r.Vars[k].Value
	}
	return result
}

// UnmanagedRefs returns key → secret ref for unmanaged secrets.
func (r layeredEnvResult) UnmanagedRefs() map[string]string {
	result := map[string]string{}
	for _, k := range r.Unmanaged {
		if ref := r.Vars[k].SecretRef; ref != nil {
			result[k] = *ref
		}
	}
	return result
}

// PlatformData builds the env v1 platform data record and validates it with platformdata.ParseEnvV1.
// No secret value is ever carried in the record.
func (r layeredEnvResult) PlatformData() (string, error) {
	record := platformdata.EnvV1{
		Platform:  r.Platform.Name,
		Variables: map[string]platformdata.EnvV1Variable{},
	}
	for k, ev := range r.Vars {
		meta := r.Keys[k]
		v := platformdata.EnvV1Variable{
			Template:   meta.Template,
			Source:     meta.Source,
			Capability: meta.Capability,
		}
		switch {
		case ev.SecretRef != nil:
			v.Sensitive = true
			v.Ref = &platformdata.EnvV1Ref{Type: platformdata.RefTypeSecret, Id: *ev.SecretRef}
		case ev.IsSensitive:
			v.Sensitive = true
		case ev.FieldRef != nil:
			v.Ref = &platformdata.EnvV1Ref{Type: platformdata.RefTypeK8sField, ApiVersion: ev.FieldRef.ApiVersion, FieldPath: ev.FieldRef.FieldPath}
		case ev.ConfigMapRef != nil:
			v.Ref = &platformdata.EnvV1Ref{Type: platformdata.RefTypeK8sConfigMap, Name: ev.ConfigMapRef.Name, Key: ev.ConfigMapRef.Key, Optional: ev.ConfigMapRef.Optional}
		case ev.ResourceFieldRef != nil:
			v.Ref = &platformdata.EnvV1Ref{Type: platformdata.RefTypeK8sResourceField, Resource: ev.ResourceFieldRef.Resource, Container: ev.ResourceFieldRef.Container, Divisor: ev.ResourceFieldRef.Divisor}
		case ev.FileKeyRef != nil:
			v.Ref = &platformdata.EnvV1Ref{Type: platformdata.RefTypeK8sFileKey, VolumeName: ev.FileKeyRef.VolumeName, Path: ev.FileKeyRef.Path, Key: ev.FileKeyRef.Key}
		default:
			v.Value = ev.Value
		}
		if meta.SecretInput {
			// The template of a secrets input is the secret value itself; never carry it.
			v.Template = ""
		}
		record.Variables[k] = v
	}

	raw, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("unable to encode env platform data: %w", err)
	}
	if _, err := platformdata.ParseEnvV1(raw); err != nil {
		return "", fmt.Errorf("generated env platform data is invalid: %w", err)
	}
	return string(raw), nil
}

// resolveLayers folds the layers (standard < cloud < otel < capability < user), interpolates,
// applies platform gates, and classifies every key.
func resolveLayers(in layeredEnvInput) (layeredEnvResult, []*tfprotov5.Diagnostic) {
	result := layeredEnvResult{Keys: map[string]layeredEnvKey{}}
	var diags []*tfprotov5.Diagnostic
	addErr := func(summary, detail string) {
		diags = append(diags, &tfprotov5.Diagnostic{Severity: tfprotov5.DiagnosticSeverityError, Summary: summary, Detail: detail})
	}

	platform, ok := platformdata.LookupPlatform(in.Platform)
	if !ok {
		addErr(fmt.Sprintf("Unknown platform: %s", in.Platform),
			fmt.Sprintf("platform must be one of: %s", strings.Join(platformdata.Platforms(), ", ")))
		return result, diags
	}
	result.Platform = platform

	type folded struct {
		Value    string
		IsSecret bool
	}
	values := map[string]folded{}
	set := func(attr, key, value string, isSecret bool, source, capability string) {
		if !validEnvVariableKey(key) {
			addErr(fmt.Sprintf("Invalid environment variable key: %s", key), invalidEnvKeyDetail)
			return
		}
		if isSecret && hasRuntimeRefTemplate(value) {
			// D16a: a secrets input holds the secret value itself; it cannot point at a runtime reference.
			addErr(fmt.Sprintf("Invalid secret template: %s", key), invalidSecretTemplateDetail(attr, key))
			return
		}
		// D16: secrets always win. Once any secrets layer sets a key, a plain layer at the same key
		// never overrides it, regardless of layer order; secret layers still override each other in order.
		if prev, exists := values[key]; exists && prev.IsSecret && !isSecret {
			return
		}
		values[key] = folded{Value: value, IsSecret: isSecret}
		result.Keys[key] = layeredEnvKey{Template: value, Source: source, Capability: capability, SecretInput: isSecret}
	}
	setMap := func(attr string, m map[string]string, isSecret bool, source string) {
		for _, k := range sortedKeys(m) {
			set(attr, k, m[k], isSecret, source, "")
		}
	}
	setCapability := func(attr string, entries []capabilityEntry, isSecret bool) {
		seen := map[string]string{}
		for _, e := range entries {
			if diag := validateCapabilityEntry(attr, e.Capability, in.CapabilityPrefixes); diag != nil {
				diags = append(diags, diag)
				continue
			}
			key := in.CapabilityPrefixes[e.Capability] + e.Name
			if prev, dup := seen[key]; dup {
				addErr(fmt.Sprintf("Duplicate capability environment variable: %s", key),
					fmt.Sprintf("%s contains %q from capabilities %q and %q; the final key (prefix + name) must be unique.", attr, key, prev, e.Capability))
				continue
			}
			seen[key] = e.Capability
			set(attr, key, e.Value, isSecret, platformdata.SourceCapability, e.Capability)
		}
	}

	setMap("standard", in.Standard, false, platformdata.SourceStandard)
	setMap("cloud", in.Cloud, false, platformdata.SourceCloud)
	setMap("otel", in.Otel, false, platformdata.SourceOtel)
	setCapability("capability_env", in.CapabilityEnv, false)
	setCapability("capability_secrets", in.CapabilitySecrets, true)
	setMap("user_env", in.UserEnv, false, platformdata.SourceUser)
	setMap("user_secrets", in.UserSecrets, true, platformdata.SourceUser)
	if len(diags) > 0 {
		return result, diags
	}

	envVars, secrets := map[string]string{}, map[string]string{}
	for k, v := range values {
		if v.IsSecret {
			secrets[k] = v.Value
		} else {
			envVars[k] = v.Value
		}
	}
	ev := NewEnvVars(envVars, secrets)
	if errs := ev.Interpolate(); len(errs) > 0 {
		for _, err := range errs {
			addErr("Invalid environment variable template", err.Error())
		}
		return result, diags
	}
	result.Vars = ev

	// Platform gates + classification
	for _, k := range sortedKeys(ev) {
		v := ev[k]
		hasK8sRef := v.FieldRef != nil || v.ConfigMapRef != nil || v.ResourceFieldRef != nil || v.FileKeyRef != nil
		switch {
		case v.SecretRef != nil:
			if !platform.SupportsSecretRefs {
				addErr(fmt.Sprintf("Unsupported template for platform %s: %s", platform.Name, k),
					fmt.Sprintf("%q uses `{{ secret(...) }}`, which is not supported on the %q platform.", k, platform.Name))
				continue
			}
			result.Unmanaged = append(result.Unmanaged, k)
		case v.IsSensitive:
			result.Managed = append(result.Managed, k)
		case hasK8sRef:
			if !platform.SupportsK8sRefs {
				addErr(fmt.Sprintf("Unsupported template for platform %s: %s", platform.Name, k),
					fmt.Sprintf("%q uses a `{{ k8s.*(...) }}` template, which is only supported on Kubernetes platforms (not %q).", k, platform.Name))
				continue
			}
			result.K8s = append(result.K8s, k)
		default:
			result.Plain = append(result.Plain, k)
		}
	}
	return result, diags
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// --- tftypes helpers ---

// capabilityEntriesFromTfValue reads a list(object({capability, name[, value]})) value.
// Null/unknown lists yield no entries; null/unknown attributes yield "".
func capabilityEntriesFromTfValue(tfVal tftypes.Value) []capabilityEntry {
	result := make([]capabilityEntry, 0)
	if tfVal.IsNull() || !tfVal.IsKnown() {
		return result
	}
	elems := make([]tftypes.Value, 0)
	if err := tfVal.As(&elems); err != nil {
		return result
	}
	for _, elem := range elems {
		if elem.IsNull() || !elem.IsKnown() {
			continue
		}
		attrs := map[string]tftypes.Value{}
		if err := elem.As(&attrs); err != nil {
			continue
		}
		result = append(result, capabilityEntry{
			Capability: extractStringFromTfValue(attrs["capability"]),
			Name:       extractStringFromTfValue(attrs["name"]),
			Value:      extractStringFromTfValue(attrs["value"]),
		})
	}
	return result
}

// echoInput returns a typed null when the config value is null (so echoed inputs keep their type),
// otherwise the config value itself.
func echoInput(tfVal tftypes.Value, typ tftypes.Type) tftypes.Value {
	if tfVal.IsNull() {
		return tftypes.NewValue(typ, nil)
	}
	return tfVal
}

// invalidSecretTemplateDetail explains why a secrets input cannot carry a runtime reference template (D16a).
func invalidSecretTemplateDetail(attr, key string) string {
	return fmt.Sprintf("%s.%s uses a `{{ secret(...) }}` or `{{ k8s.*(...) }}` template. Secrets hold the secret value itself and cannot use runtime reference templates; move %s to env_vars (user_env) instead.", attr, key, key)
}

// validateCapabilityEntry applies D16b/D16c to one capability list entry: the capability name is required and
// must have a prefix in capability_prefixes (an explicit "" prefix is fine).
func validateCapabilityEntry(attr, capability string, prefixes map[string]string) *tfprotov5.Diagnostic {
	if capability == "" {
		return &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Capability name is required",
			Detail:   fmt.Sprintf("%s contains an entry with an empty capability; every entry must name the capability that emitted it.", attr),
		}
	}
	if _, ok := prefixes[capability]; !ok {
		return &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("Unknown capability: %s", capability),
			Detail:   fmt.Sprintf("%s references capability %q, which has no entry in capability_prefixes. Every capability must appear in capability_prefixes (use \"\" for no prefix).", attr, capability),
		}
	}
	return nil
}

// keysToEmptyMap builds a map with "" values from a key set, so key-only inputs (ns_env_layout, D15)
// feed the same resolver as full-value inputs.
func keysToEmptyMap(keys []string) map[string]string {
	result := make(map[string]string, len(keys))
	for _, k := range keys {
		result[k] = ""
	}
	return result
}

// layeredEnvValuesInputFromConfig reads the full-value inputs of ns_env_values.
func layeredEnvValuesInputFromConfig(config map[string]tftypes.Value) layeredEnvInput {
	return layeredEnvInput{
		Platform:           extractStringFromConfig(config, "platform"),
		Standard:           TfValueToMap(config["standard"]),
		Cloud:              TfValueToMap(config["cloud"]),
		Otel:               TfValueToMap(config["otel"]),
		CapabilityEnv:      capabilityEntriesFromTfValue(config["capability_env"]),
		CapabilitySecrets:  capabilityEntriesFromTfValue(config["capability_secrets"]),
		CapabilityPrefixes: TfValueToMap(config["capability_prefixes"]),
		UserEnv:            TfValueToMap(config["user_env"]),
		UserSecrets:        TfValueToMap(config["user_secrets"]),
	}
}

// layeredEnvLayoutInputFromConfig reads the key-only inputs of ns_env_layout (D15) into the same shape,
// using "" as the value of every key-only layer. `user_env` keeps its templates so secret promotion and
// `{{ secret(...) }}` refs are detected.
func layeredEnvLayoutInputFromConfig(config map[string]tftypes.Value) layeredEnvInput {
	return layeredEnvInput{
		Platform:           extractStringFromConfig(config, "platform"),
		Standard:           keysToEmptyMap(TfSetValueToStringSlice(config["standard_keys"])),
		Cloud:              keysToEmptyMap(TfSetValueToStringSlice(config["cloud_keys"])),
		Otel:               keysToEmptyMap(TfSetValueToStringSlice(config["otel_keys"])),
		CapabilityEnv:      capabilityEntriesFromTfValue(config["capability_env_keys"]),
		CapabilitySecrets:  capabilityEntriesFromTfValue(config["capability_secret_keys"]),
		CapabilityPrefixes: TfValueToMap(config["capability_prefixes"]),
		UserEnv:            TfValueToMap(config["user_env"]),
		UserSecrets:        keysToEmptyMap(TfSetValueToStringSlice(config["user_secret_keys"])),
	}
}

// layeredEnvValidation names the config attributes validateLayeredEnv checks, by shape.
type layeredEnvValidation struct {
	// MapAttrs are map(string) attributes whose keys are validated.
	MapAttrs []string
	// SetAttrs are set(string) attributes whose elements are validated as keys.
	SetAttrs []string
	// CapabilityAttrs are list(object({capability, name[, value]})) attributes: capability checks (D16b/c) + key validation.
	CapabilityAttrs []string
	// SecretMapAttrs are map(string) secrets whose values must not be runtime reference templates (D16a).
	SecretMapAttrs []string
	// SecretCapabilityAttrs are capability secret lists: capability checks + key validation + D16a on values.
	SecretCapabilityAttrs []string
}

// validateLayeredEnv validates everything that is known at validation time.
// Unknown values are skipped; Terraform calls Read once everything is known, and Read repeats every check.
func validateLayeredEnv(config map[string]tftypes.Value, v layeredEnvValidation) []*tfprotov5.Diagnostic {
	diags := make([]*tfprotov5.Diagnostic, 0)
	checkKey := func(key string) {
		if key == "" || !validEnvVariableKey(key) {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Invalid environment variable key: %s", key),
				Detail:   invalidEnvKeyDetail,
			})
		}
	}
	checkSecretValue := func(attr, key string, val tftypes.Value) {
		if val.IsNull() || !val.IsKnown() {
			return
		}
		if hasRuntimeRefTemplate(extractStringFromTfValue(val)) {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Invalid secret template: %s", key),
				Detail:   invalidSecretTemplateDetail(attr, key),
			})
		}
	}
	// A null capability_prefixes is an empty map (every capability is then unknown); a not-fully-known one
	// defers all capability checks to Read.
	prefixesKnown := config["capability_prefixes"].IsFullyKnown()
	prefixes := map[string]string{}
	if prefixesKnown && !config["capability_prefixes"].IsNull() {
		prefixes = TfValueToMap(config["capability_prefixes"])
	}

	if pv := config["platform"]; !pv.IsNull() && pv.IsKnown() {
		name := extractStringFromTfValue(pv)
		if _, ok := platformdata.LookupPlatform(name); !ok {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Unknown platform: %s", name),
				Detail:   fmt.Sprintf("platform must be one of: %s", strings.Join(platformdata.Platforms(), ", ")),
			})
		}
	}
	for _, attr := range v.MapAttrs {
		mv := config[attr]
		if mv.IsNull() || !mv.IsKnown() {
			continue
		}
		for key := range TfValueToMap(mv) {
			checkKey(key)
		}
	}
	for _, attr := range v.SetAttrs {
		sv := config[attr]
		if sv.IsNull() || !sv.IsFullyKnown() {
			continue
		}
		for _, key := range TfSetValueToStringSlice(sv) {
			checkKey(key)
		}
	}
	for _, attr := range v.SecretMapAttrs {
		mv := config[attr]
		if mv.IsNull() || !mv.IsKnown() {
			continue
		}
		elems := map[string]tftypes.Value{}
		if err := mv.As(&elems); err != nil {
			continue
		}
		for key, val := range elems {
			checkSecretValue(attr, key, val)
		}
	}
	isSecretList := map[string]bool{}
	for _, attr := range v.SecretCapabilityAttrs {
		isSecretList[attr] = true
	}
	capabilityAttrs := append(append([]string{}, v.CapabilityAttrs...), v.SecretCapabilityAttrs...)
	for _, attr := range capabilityAttrs {
		lv := config[attr]
		if lv.IsNull() || !lv.IsKnown() {
			continue
		}
		elems := make([]tftypes.Value, 0)
		if err := lv.As(&elems); err != nil {
			continue
		}
		for _, elem := range elems {
			if elem.IsNull() || !elem.IsKnown() {
				continue
			}
			attrs := map[string]tftypes.Value{}
			if err := elem.As(&attrs); err != nil {
				continue
			}
			nameVal, capVal := attrs["name"], attrs["capability"]
			if !capVal.IsKnown() || !prefixesKnown {
				continue
			}
			capability := extractStringFromTfValue(capVal)
			if diag := validateCapabilityEntry(attr, capability, prefixes); diag != nil {
				diags = append(diags, diag)
				continue
			}
			if nameVal.IsNull() || !nameVal.IsKnown() {
				continue
			}
			key := prefixes[capability] + extractStringFromTfValue(nameVal)
			checkKey(key)
			if isSecretList[attr] {
				checkSecretValue(attr, key, attrs["value"])
			}
		}
	}
	return diags
}
