package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type dataEnvLayout struct {
	p *provider
}

func newDataEnvLayout(p *provider) (*dataEnvLayout, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataEnvLayout{p: p}, nil
}

// layeredEnvPlatformAttr is the `platform` input common to ns_env_layout and ns_env_values.
func layeredEnvPlatformAttr() *tfprotov5.SchemaAttribute {
	return &tfprotov5.SchemaAttribute{
		Name:            "platform",
		Type:            tftypes.String,
		Description:     "The runtime platform of the application (e.g. `aws_ecs`, `gcp_gke`, `aws_lambda`, `gcp_cloudrun`; always cloud-prefixed). Gates which templates are allowed: `{{ secret(...) }}` and `{{ k8s.*(...) }}`.",
		DescriptionKind: tfprotov5.StringKindMarkdown,
		Required:        true,
	}
}

// layeredEnvSharedInputAttrs are the input attributes common to ns_env_layout and ns_env_values
// beyond `platform`. The layout takes the non-user layers as keys only (D15), so those are defined per data source.
func layeredEnvSharedInputAttrs() []*tfprotov5.SchemaAttribute {
	return []*tfprotov5.SchemaAttribute{
		{
			Name:            "capability_prefixes",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Map of capability `capability` to the prefix applied to its environment variable and secret names. Every capability referenced by a capability list must appear here (use `\"\"` for no prefix).",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "user_env",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "User-defined environment variables (`var.env_vars`). Highest precedence among plain layers; a key that any secrets input also sets stays a secret.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
	}
}

// layeredEnvSharedOutputAttrs are the computed attributes common to ns_env_layout and ns_env_values.
func layeredEnvSharedOutputAttrs() []*tfprotov5.SchemaAttribute {
	return []*tfprotov5.SchemaAttribute{
		{
			Name:            "managed_secret_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "Keys of secrets whose cloud secret the module must create (secrets inputs and variables promoted by interpolating a secret).",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "unmanaged_secret_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "Keys of secrets that reference an existing cloud secret via `{{ secret(...) }}`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "all_secret_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "Union of `managed_secret_keys` and `unmanaged_secret_keys`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "sources",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Map of every key to the layer that supplied it: `standard`, `cloud`, `otel`, `capability`, or `user`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "capabilities",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Map of capability-sourced keys to the `capability` of the capability that supplied them.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
	}
}

func (*dataEnvLayout) Schema(ctx context.Context) *tfprotov5.Schema {
	attrs := []*tfprotov5.SchemaAttribute{
		deprecatedIDAttribute(),
		layeredEnvPlatformAttr(),
		// D15: non-user layers are keys only so the layout is always resolvable at plan.
		{
			Name:            "standard_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "Keys of the standard Nullstone environment variables (`keys(local.standard_env_vars)`). Lowest precedence.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "cloud_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "Keys of the cloud platform environment variables (e.g. `AWS_REGION`, `GOOGLE_CLOUD_PROJECT`). Overrides `standard_keys`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "otel_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "Keys of the OpenTelemetry environment variables (`OTEL_*`). Overrides `cloud_keys`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "capability_env_keys",
			Type:            tftypes.List{ElementType: capabilityKeyEntryType},
			Description:     "Environment variable keys emitted by capabilities. Each entry is `{ capability, name }`; the final key is `capability_prefixes[capability] + name`. Overrides `otel_keys`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
		{
			Name:            "capability_secret_keys",
			Type:            tftypes.List{ElementType: capabilityKeyEntryType},
			Description:     "Secret keys emitted by capabilities. Each entry is `{ capability, name }`; the final key is `capability_prefixes[capability] + name`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
	}
	attrs = append(attrs, layeredEnvSharedInputAttrs()...)
	attrs = append(attrs,
		&tfprotov5.SchemaAttribute{
			Name:            "user_secret_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "User-defined secret keys (`nonsensitive(keys(var.secrets))`).",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Optional:        true,
		},
	)
	attrs = append(attrs, layeredEnvSharedOutputAttrs()...)

	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description: "Data source that merges the layers of an application's environment (standard, cloud, otel, capability, user) using keys only " +
				"(plus `user_env` templates), so that the set of secrets the module must create is known at plan time.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes:      attrs,
		},
	}
}

func (d *dataEnvLayout) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := validateLayeredEnv(config, layeredEnvValidation{
		MapAttrs:        []string{"user_env"},
		SetAttrs:        []string{"standard_keys", "cloud_keys", "otel_keys", "user_secret_keys"},
		CapabilityAttrs: []string{"capability_env_keys", "capability_secret_keys"},
	})
	return diags, nil
}

func (d *dataEnvLayout) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	result, diags := resolveLayers(layeredEnvLayoutInputFromConfig(config))
	if hasErrorDiagnostic(diags) {
		return nil, diags, nil
	}

	id := result.Vars.KeysHash()
	tflog.Debug(ctx, "id", map[string]interface{}{"value": id})
	tflog.Debug(ctx, "managed_secret_keys", map[string]interface{}{"value": result.Managed})
	tflog.Debug(ctx, "unmanaged_secret_keys", map[string]interface{}{"value": result.Unmanaged})

	stringSet := tftypes.Set{ElementType: tftypes.String}
	return map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, id),
		"platform":               config["platform"],
		"standard_keys":          echoInput(config["standard_keys"], stringSet),
		"cloud_keys":             echoInput(config["cloud_keys"], stringSet),
		"otel_keys":              echoInput(config["otel_keys"], stringSet),
		"capability_env_keys":    echoInput(config["capability_env_keys"], tftypes.List{ElementType: capabilityKeyEntryType}),
		"capability_secret_keys": echoInput(config["capability_secret_keys"], tftypes.List{ElementType: capabilityKeyEntryType}),
		"capability_prefixes":    echoInput(config["capability_prefixes"], tftypes.Map{ElementType: tftypes.String}),
		"user_env":               echoInput(config["user_env"], tftypes.Map{ElementType: tftypes.String}),
		"user_secret_keys":       echoInput(config["user_secret_keys"], stringSet),
		"managed_secret_keys":    SliceToTfSet(result.Managed),
		"unmanaged_secret_keys":  SliceToTfSet(result.Unmanaged),
		"all_secret_keys":        SliceToTfSet(result.AllSecretKeys()),
		"sources":                MapToTfValue(result.Sources()),
		"capabilities":           MapToTfValue(result.Capabilities()),
	}, diags, nil
}
