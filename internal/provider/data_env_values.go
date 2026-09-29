package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type dataEnvValues struct {
	p *provider
}

func newDataEnvValues(p *provider) (*dataEnvValues, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataEnvValues{p: p}, nil
}

func (*dataEnvValues) Schema(ctx context.Context) *tfprotov5.Schema {
	attrs := []*tfprotov5.SchemaAttribute{
		deprecatedIDAttribute(),
		layeredEnvPlatformAttr(),
		{
			Name:            "standard",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Standard Nullstone environment variables (`NULLSTONE_*`). Lowest precedence.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "cloud",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Cloud platform environment variables (e.g. `AWS_REGION`, `GOOGLE_CLOUD_PROJECT`). Overrides `standard`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "otel",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "OpenTelemetry environment variables (`OTEL_*`). Overrides `cloud`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "capability_env",
			Type:            tftypes.List{ElementType: capabilityEnvEntryType},
			Description:     "Environment variables emitted by capabilities (`local.capabilities.env`). Each entry is `{ capability, name, value }`; the final key is `capability_prefixes[capability] + name`. Overrides `otel`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
	}
	attrs = append(attrs, layeredEnvSharedInputAttrs()...)
	attrs = append(attrs,
		&tfprotov5.SchemaAttribute{
			Name:            "capability_secrets",
			Type:            tftypes.List{ElementType: capabilityEnvEntryType},
			Description:     "Secrets emitted by capabilities (`local.capabilities.secrets`). Each entry is `{ capability, name, value }`; the final key is `capability_prefixes[capability] + name`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
			Sensitive:       true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "user_secrets",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "User-defined secrets (`var.secrets`). Highest precedence.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
			Sensitive:       true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "env_variables",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Plain environment variables after interpolation. Secrets and template refs are excluded.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "secrets",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Managed secrets after interpolation (the module creates a cloud secret for each).",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
			Sensitive:       true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "unmanaged_secret_refs",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Map of environment variables that refer to an existing secret (from `{{ secret(...) }}`) to their secret reference.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "field_refs",
			Type:            tftypes.Map{ElementType: fieldRefObjectType},
			Description:     "Map of environment variables that refer to a Kubernetes field for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "config_map_refs",
			Type:            tftypes.Map{ElementType: configMapRefObjectType},
			Description:     "Map of environment variables that refer to a Kubernetes ConfigMap key for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "resource_field_refs",
			Type:            tftypes.Map{ElementType: resourceFieldRefObjectType},
			Description:     "Map of environment variables that refer to a Kubernetes resource field for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "file_key_refs",
			Type:            tftypes.Map{ElementType: fileKeyRefObjectType},
			Description:     "Map of environment variables that refer to a Kubernetes file key for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		&tfprotov5.SchemaAttribute{
			Name:            "platform_data",
			Type:            tftypes.String,
			Description:     "JSON-encoded `env` (version 1) platform data record describing every variable, its source, and whether it is sensitive. Never contains secret values. Pass to `ns_platform_data` with `kind = \"env\"` and `version = 1`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
	)
	attrs = append(attrs, layeredEnvSharedOutputAttrs()...)

	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description: "Data source that merges the layers of an application's environment (standard, cloud, otel, capability, user), " +
				"interpolates `{{ VAR }}` references, splits secrets into managed and unmanaged, extracts Kubernetes `valueFrom` refs, " +
				"and emits the `env` platform data record.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes:      attrs,
		},
	}
}

func (d *dataEnvValues) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := validateLayeredEnv(config, layeredEnvValidation{
		MapAttrs:              []string{"standard", "cloud", "otel", "user_env", "user_secrets"},
		CapabilityAttrs:       []string{"capability_env"},
		SecretMapAttrs:        []string{"user_secrets"},
		SecretCapabilityAttrs: []string{"capability_secrets"},
	})
	return diags, nil
}

func (d *dataEnvValues) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	result, diags := resolveLayers(layeredEnvValuesInputFromConfig(config))
	if hasErrorDiagnostic(diags) {
		return nil, diags, nil
	}

	platformData, err := result.PlatformData()
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to build env platform data",
			Detail:   err.Error(),
		})
		return nil, diags, nil
	}

	id := result.Vars.KeysHash()
	tflog.Debug(ctx, "id", map[string]interface{}{"value": id})
	tflog.Debug(ctx, "managed_secret_keys", map[string]interface{}{"value": result.Managed})
	tflog.Debug(ctx, "unmanaged_secret_keys", map[string]interface{}{"value": result.Unmanaged})

	return map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, id),
		"platform":              config["platform"],
		"standard":              echoInput(config["standard"], tftypes.Map{ElementType: tftypes.String}),
		"cloud":                 echoInput(config["cloud"], tftypes.Map{ElementType: tftypes.String}),
		"otel":                  echoInput(config["otel"], tftypes.Map{ElementType: tftypes.String}),
		"capability_env":        echoInput(config["capability_env"], tftypes.List{ElementType: capabilityEnvEntryType}),
		"capability_secrets":    echoInput(config["capability_secrets"], tftypes.List{ElementType: capabilityEnvEntryType}),
		"capability_prefixes":   echoInput(config["capability_prefixes"], tftypes.Map{ElementType: tftypes.String}),
		"user_env":              echoInput(config["user_env"], tftypes.Map{ElementType: tftypes.String}),
		"user_secrets":          echoInput(config["user_secrets"], tftypes.Map{ElementType: tftypes.String}),
		"env_variables":         MapToTfValue(result.PlainValues()),
		"secrets":               MapToTfValue(result.ManagedValues()),
		"unmanaged_secret_refs": MapToTfValue(result.UnmanagedRefs()),
		"field_refs":            FieldRefsToTfValue(result.Vars.FieldRefs()),
		"config_map_refs":       ConfigMapRefsToTfValue(result.Vars.ConfigMapRefs()),
		"resource_field_refs":   ResourceFieldRefsToTfValue(result.Vars.ResourceFieldRefs()),
		"file_key_refs":         FileKeyRefsToTfValue(result.Vars.FileKeyRefs()),
		"platform_data":         tftypes.NewValue(tftypes.String, platformData),
		"managed_secret_keys":   SliceToTfSet(result.Managed),
		"unmanaged_secret_keys": SliceToTfSet(result.Unmanaged),
		"all_secret_keys":       SliceToTfSet(result.AllSecretKeys()),
		"sources":               MapToTfValue(result.Sources()),
		"capabilities":          MapToTfValue(result.Capabilities()),
	}, diags, nil
}
