package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nullstone-io/module/platformdata"
)

// ns_env_platform_data (D14) completes the env v1 record emitted by `ns_env_values` with the ids of the
// managed secrets the module created, so the record Nullstone reads names every secret.
// See specs/env-layers.md.

type dataEnvPlatformData struct {
	p *provider
}

func newDataEnvPlatformData(p *provider) (*dataEnvPlatformData, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataEnvPlatformData{p: p}, nil
}

// k8sSecretRefObjectType is the element type of `k8s_secret_refs`.
var k8sSecretRefObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"name": tftypes.String,
	"key":  tftypes.String,
}}

func (*dataEnvPlatformData) Schema(ctx context.Context) *tfprotov5.Schema {
	attrs := []*tfprotov5.SchemaAttribute{
		deprecatedIDAttribute(),
		{
			Name:            "values",
			Type:            tftypes.String,
			Description:     "The env v1 platform data record produced by `ns_env_values` (`data.ns_env_values.this.platform_data`).",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Required:        true,
		},
		{
			Name:            "secret_ids",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Map of managed secret key to the id of the cloud secret the module created for it (e.g. an AWS Secrets Manager ARN or a GCP secret resource name).",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "k8s_secret_refs",
			Type:            tftypes.Map{ElementType: k8sSecretRefObjectType},
			Description:     "Map of managed secret key to the Kubernetes `secretKeyRef` (`{ name, key }`) the module created for it. Only valid on Kubernetes platforms.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "platform_data",
			Type:            tftypes.String,
			Description:     "JSON-encoded `env` (version 1) platform data record with every managed secret carrying its `secret` or `k8s_secret_key` ref. Never contains secret values. Nullstone reads this attribute directly from state.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
	}

	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description: "Data source that completes the env platform data record from `ns_env_values` with the ids of the managed secrets " +
				"the module created. Every managed secret must receive an id (`secret_ids`) or a Kubernetes secret ref (`k8s_secret_refs`).",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes:      attrs,
		},
	}
}

// k8sSecretRef is one `k8s_secret_refs` entry.
type k8sSecretRef struct {
	Name string
	Key  string
}

// envPlatformDataInput is the fully-known configuration of ns_env_platform_data.
type envPlatformDataInput struct {
	Values        string
	SecretIds     map[string]string
	K8sSecretRefs map[string]k8sSecretRef
}

func envPlatformDataInputFromConfig(config map[string]tftypes.Value) envPlatformDataInput {
	in := envPlatformDataInput{
		Values:        extractStringFromConfig(config, "values"),
		SecretIds:     TfValueToMap(config["secret_ids"]),
		K8sSecretRefs: map[string]k8sSecretRef{},
	}
	if v := config["k8s_secret_refs"]; !v.IsNull() && v.IsKnown() {
		elems := map[string]tftypes.Value{}
		if err := v.As(&elems); err == nil {
			for key, elem := range elems {
				attrs := map[string]tftypes.Value{}
				if elem.IsNull() || !elem.IsKnown() {
					continue
				}
				if err := elem.As(&attrs); err != nil {
					continue
				}
				in.K8sSecretRefs[key] = k8sSecretRef{
					Name: extractStringFromTfValue(attrs["name"]),
					Key:  extractStringFromTfValue(attrs["key"]),
				}
			}
		}
	}
	return in
}

func (d *dataEnvPlatformData) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	valuesVal := config["values"]
	if valuesVal.IsNull() || !valuesVal.IsKnown() {
		return nil, nil
	}
	// `values` is known: parse it so a malformed record fails at plan.
	// When the id maps are fully known as well, run the whole completion so coverage errors surface at plan too.
	if !config["secret_ids"].IsFullyKnown() || !config["k8s_secret_refs"].IsFullyKnown() {
		var diags []*tfprotov5.Diagnostic
		if _, err := platformdata.ParseEnvV1(json.RawMessage(extractStringFromTfValue(valuesVal))); err != nil {
			diags = append(diags, invalidEnvValuesDiag(err))
		}
		return diags, nil
	}
	_, diags := completeEnvPlatformData(envPlatformDataInputFromConfig(config))
	return diags, nil
}

func (d *dataEnvPlatformData) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	record, diags := completeEnvPlatformData(envPlatformDataInputFromConfig(config))
	if hasErrorDiagnostic(diags) {
		return nil, diags, nil
	}

	raw, err := json.Marshal(record)
	if err != nil {
		return nil, append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to build env platform data",
			Detail:   fmt.Sprintf("unable to encode env platform data: %s", err),
		}), nil
	}
	// Re-parse before emitting, like layeredEnvResult.PlatformData(): a failure here is a provider bug.
	if _, err := platformdata.ParseEnvV1(raw); err != nil {
		return nil, append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to build env platform data",
			Detail:   fmt.Sprintf("generated env platform data is invalid: %s", err),
		}), nil
	}

	// id: hash of the record's sorted keys (+ sensitivity), reusing EnvVars.KeysHash.
	keys := EnvVars{}
	for k, v := range record.Variables {
		keys[k] = EnvVar{IsSensitive: v.IsSensitive()}
	}
	id := keys.KeysHash()
	tflog.Debug(ctx, "id", map[string]interface{}{"value": id})

	return map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, id),
		"values":          config["values"],
		"secret_ids":      echoInput(config["secret_ids"], tftypes.Map{ElementType: tftypes.String}),
		"k8s_secret_refs": echoInput(config["k8s_secret_refs"], tftypes.Map{ElementType: k8sSecretRefObjectType}),
		"platform_data":   tftypes.NewValue(tftypes.String, string(raw)),
	}, diags, nil
}

// completeEnvPlatformData parses the env v1 record and attaches a `secret` or `k8s_secret_key` ref to every
// managed secret (Sensitive with no Ref). Every managed secret must receive exactly one id.
func completeEnvPlatformData(in envPlatformDataInput) (platformdata.EnvV1, []*tfprotov5.Diagnostic) {
	var diags []*tfprotov5.Diagnostic
	addErr := func(summary, detail string) {
		diags = append(diags, &tfprotov5.Diagnostic{Severity: tfprotov5.DiagnosticSeverityError, Summary: summary, Detail: detail})
	}

	record, err := platformdata.ParseEnvV1(json.RawMessage(in.Values))
	if err != nil {
		return platformdata.EnvV1{}, []*tfprotov5.Diagnostic{invalidEnvValuesDiag(err)}
	}

	// A managed secret is Sensitive with no Ref: ns_env_values leaves the ref for the module to mint.
	isManagedSecret := func(key string) (bool, string) {
		v, ok := record.Variables[key]
		if !ok {
			return false, fmt.Sprintf("%q is not a variable in values.", key)
		}
		if !v.Sensitive || v.Ref != nil {
			return false, fmt.Sprintf("%q is not a managed secret; only keys in managed_secret_keys take an id.", key)
		}
		return true, ""
	}

	for _, key := range sortedKeys(in.SecretIds) {
		if _, both := in.K8sSecretRefs[key]; both {
			addErr(fmt.Sprintf("Secret id declared twice: %s", key),
				fmt.Sprintf("%q appears in both secret_ids and k8s_secret_refs; a managed secret takes exactly one id.", key))
			continue
		}
		ok, why := isManagedSecret(key)
		if !ok {
			addErr(fmt.Sprintf("Invalid secret id key: %s", key), fmt.Sprintf("secret_ids: %s", why))
			continue
		}
		v := record.Variables[key]
		v.Ref = &platformdata.EnvV1Ref{Type: platformdata.RefTypeSecret, Id: in.SecretIds[key]}
		record.Variables[key] = v
	}

	for _, key := range sortedKeys(in.K8sSecretRefs) {
		if _, both := in.SecretIds[key]; both {
			continue // already reported above
		}
		ok, why := isManagedSecret(key)
		if !ok {
			addErr(fmt.Sprintf("Invalid secret id key: %s", key), fmt.Sprintf("k8s_secret_refs: %s", why))
			continue
		}
		if platform, known := platformdata.LookupPlatform(record.Platform); known && !platform.SupportsK8sRefs {
			addErr(fmt.Sprintf("Unsupported secret ref for platform %s: %s", platform.Name, key),
				fmt.Sprintf("k8s_secret_refs.%s is a Kubernetes secretKeyRef, which is only supported on Kubernetes platforms (not %q); use secret_ids instead.", key, platform.Name))
			continue
		}
		ref := in.K8sSecretRefs[key]
		v := record.Variables[key]
		v.Ref = &platformdata.EnvV1Ref{Type: platformdata.RefTypeK8sSecretKey, Name: ref.Name, Key: ref.Key}
		record.Variables[key] = v
	}

	// Coverage: every managed secret must now carry a ref.
	missing := make([]string, 0)
	for _, key := range record.Keys() {
		if v := record.Variables[key]; v.Sensitive && v.Ref == nil {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		addErr(fmt.Sprintf("Managed secrets without an id: %s", strings.Join(missing, ", ")),
			"Every key in managed_secret_keys must appear in secret_ids (cloud secret id) or k8s_secret_refs (Kubernetes secretKeyRef) so Nullstone knows where each managed secret lives.")
	}

	return record, diags
}

func invalidEnvValuesDiag(err error) *tfprotov5.Diagnostic {
	return &tfprotov5.Diagnostic{
		Severity: tfprotov5.DiagnosticSeverityError,
		Summary:  "Invalid env platform data",
		Detail:   fmt.Sprintf("values must be the env v1 record from ns_env_values (`data.ns_env_values.<name>.platform_data`): %s", err),
	}
}
