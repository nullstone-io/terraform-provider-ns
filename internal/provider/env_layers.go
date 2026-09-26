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
	"cap_tf_id": tftypes.String,
	"name":      tftypes.String,
	"value":     tftypes.String,
}}

// capabilityKeyEntryType is the element type of `capability_secret_keys` (no value).
var capabilityKeyEntryType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"cap_tf_id": tftypes.String,
	"name":      tftypes.String,
}}

const invalidEnvKeyDetail = "An environment variable key can only contain letters, numbers, and the underscore character. It also can not begin with a number."

// capabilityEntry is one env var or secret emitted by a capability module.
type capabilityEntry struct {
	CapTfId string
	Name    string
	Value   string
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
	set := func(key, value string, isSecret bool, source, capability string) {
		if !validEnvVariableKey(key) {
			addErr(fmt.Sprintf("Invalid environment variable key: %s", key), invalidEnvKeyDetail)
			return
		}
		values[key] = folded{Value: value, IsSecret: isSecret}
		result.Keys[key] = layeredEnvKey{Template: value, Source: source, Capability: capability, SecretInput: isSecret}
	}
	setMap := func(m map[string]string, isSecret bool, source string) {
		for _, k := range sortedKeys(m) {
			set(k, m[k], isSecret, source, "")
		}
	}
	setCapability := func(attr string, entries []capabilityEntry, isSecret bool) {
		seen := map[string]string{}
		for _, e := range entries {
			key := in.CapabilityPrefixes[e.CapTfId] + e.Name
			if prev, dup := seen[key]; dup {
				addErr(fmt.Sprintf("Duplicate capability environment variable: %s", key),
					fmt.Sprintf("%s contains %q from capabilities %q and %q; the final key (prefix + name) must be unique.", attr, key, prev, e.CapTfId))
				continue
			}
			seen[key] = e.CapTfId
			set(key, e.Value, isSecret, platformdata.SourceCapability, e.CapTfId)
		}
	}

	setMap(in.Standard, false, platformdata.SourceStandard)
	setMap(in.Cloud, false, platformdata.SourceCloud)
	setMap(in.Otel, false, platformdata.SourceOtel)
	setCapability("capability_env", in.CapabilityEnv, false)
	setCapability("capability_secrets", in.CapabilitySecrets, true)
	setMap(in.UserEnv, false, platformdata.SourceUser)
	setMap(in.UserSecrets, true, platformdata.SourceUser)
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

// capabilityEntriesFromTfValue reads a list(object({cap_tf_id, name[, value]})) value.
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
			CapTfId: extractStringFromTfValue(attrs["cap_tf_id"]),
			Name:    extractStringFromTfValue(attrs["name"]),
			Value:   extractStringFromTfValue(attrs["value"]),
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

// layeredEnvInputFromConfig reads the shared attributes. Secrets are read by the caller since they differ per data source.
func layeredEnvInputFromConfig(config map[string]tftypes.Value) layeredEnvInput {
	return layeredEnvInput{
		Platform:           extractStringFromConfig(config, "platform"),
		Standard:           TfValueToMap(config["standard"]),
		Cloud:              TfValueToMap(config["cloud"]),
		Otel:               TfValueToMap(config["otel"]),
		CapabilityEnv:      capabilityEntriesFromTfValue(config["capability_env"]),
		CapabilityPrefixes: TfValueToMap(config["capability_prefixes"]),
		UserEnv:            TfValueToMap(config["user_env"]),
	}
}

// validateLayeredEnvKeys validates every key that is known at validation time.
// Unknown values are skipped; Terraform calls Read once everything is known.
func validateLayeredEnvKeys(config map[string]tftypes.Value, mapAttrs []string, listAttrs []string, setAttrs []string) []*tfprotov5.Diagnostic {
	diags := make([]*tfprotov5.Diagnostic, 0)
	check := func(key string) {
		if key == "" || !validEnvVariableKey(key) {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Invalid environment variable key: %s", key),
				Detail:   invalidEnvKeyDetail,
			})
		}
	}
	prefixes := map[string]string{}
	if v := config["capability_prefixes"]; !v.IsNull() && v.IsFullyKnown() {
		prefixes = TfValueToMap(v)
	}

	if v := config["platform"]; !v.IsNull() && v.IsKnown() {
		name := extractStringFromTfValue(v)
		if _, ok := platformdata.LookupPlatform(name); !ok {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Unknown platform: %s", name),
				Detail:   fmt.Sprintf("platform must be one of: %s", strings.Join(platformdata.Platforms(), ", ")),
			})
		}
	}
	for _, attr := range mapAttrs {
		v := config[attr]
		if v.IsNull() || !v.IsKnown() {
			continue
		}
		for key := range TfValueToMap(v) {
			check(key)
		}
	}
	for _, attr := range setAttrs {
		v := config[attr]
		if v.IsNull() || !v.IsFullyKnown() {
			continue
		}
		for _, key := range TfSetValueToStringSlice(v) {
			check(key)
		}
	}
	for _, attr := range listAttrs {
		v := config[attr]
		if v.IsNull() || !v.IsKnown() {
			continue
		}
		elems := make([]tftypes.Value, 0)
		if err := v.As(&elems); err != nil {
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
			nameVal, capVal := attrs["name"], attrs["cap_tf_id"]
			if nameVal.IsNull() || !nameVal.IsKnown() || !capVal.IsKnown() {
				continue
			}
			prefix, ok := prefixes[extractStringFromTfValue(capVal)]
			if !ok && !config["capability_prefixes"].IsNull() && !config["capability_prefixes"].IsFullyKnown() {
				// prefix unknown at plan; validate at read
				continue
			}
			check(prefix + extractStringFromTfValue(nameVal))
		}
	}
	return diags
}
