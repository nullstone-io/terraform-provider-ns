package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type dataEnvVariables struct {
	p *provider
}

func newDataEnvVariables(p *provider) (*dataEnvVariables, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataEnvVariables{p: p}, nil
}

func (*dataEnvVariables) Schema(ctx context.Context) *tfprotov5.Schema {
	attrs := []*tfprotov5.SchemaAttribute{
		deprecatedIDAttribute(),
		{
			Name:            "input_env_variables",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "The raw environment variables before they are interpolated.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Required:        true,
		},
		{
			Name:            "input_secrets",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "The raw secrets before they are interpolated.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Required:        true,
			Sensitive:       true,
		},
		{
			Name:            "env_variables",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "The processed environment variables after they are interpolated.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "secrets",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "The processed secrets after they are interpolated.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
			Sensitive:       true,
		},
		{
			Name:            "secret_refs",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "Map of environment variables that refer to an existing secret key for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "field_refs",
			Type:            tftypes.Map{ElementType: fieldRefObjectType},
			Description:     "Map of environment variables that refer to a Kubernetes field for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "config_map_refs",
			Type:            tftypes.Map{ElementType: configMapRefObjectType},
			Description:     "Map of environment variables that refer to a Kubernetes ConfigMap key for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "resource_field_refs",
			Type:            tftypes.Map{ElementType: resourceFieldRefObjectType},
			Description:     "Map of environment variables that refer to a Kubernetes resource field for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "file_key_refs",
			Type:            tftypes.Map{ElementType: fileKeyRefObjectType},
			Description:     "Map of environment variables that refer to a Kubernetes file key for their values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
	}

	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to interpolate any variables or env variables into their final values.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes:      attrs,
		},
	}
}

func (d *dataEnvVariables) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	inputEnvVariables := TfValueToMap(config["input_env_variables"])
	inputSecrets := TfValueToMap(config["input_secrets"])

	errors := make([]*tfprotov5.Diagnostic, 0)
	for key, _ := range inputEnvVariables {
		if !validEnvVariableKey(key) {
			errors = append(errors, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Invalid environment variable key: %s", key),
				Detail:   "An environment variable key can only contain letters, numbers, and the underscore character. It also can not begin with a number.",
			})
		}
	}
	for key, _ := range inputSecrets {
		if !validEnvVariableKey(key) {
			errors = append(errors, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Invalid environment variable key: %s", key),
				Detail:   "An environment variable key can only contain letters, numbers, and the underscore character. It also can not begin with a number.",
			})
		}
	}

	return errors, nil
}

func (d *dataEnvVariables) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	inputEnvVariables := config["input_env_variables"]
	inputSecrets := config["input_secrets"]

	tflog.Debug(ctx, "input_env_variables", inputEnvVariables)
	tflog.Debug(ctx, "input_secrets", inputSecrets)

	ev := NewEnvVars(TfValueToMap(inputEnvVariables), TfValueToMap(inputSecrets))
	if errs := ev.Interpolate(); len(errs) > 0 {
		diags := make([]*tfprotov5.Diagnostic, 0, len(errs))
		for _, err := range errs {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Invalid environment variable template",
				Detail:   err.Error(),
			})
		}
		return nil, diags, nil
	}

	// calculate the unique id for this data source based on a hash of the resulting env variables and secrets
	id := ev.Hash()
	envVariables := ev.EnvVars()
	secrets := ev.Secrets()
	secretRefs := ev.SecretRefs()

	tflog.Debug(ctx, "id", id)
	tflog.Debug(ctx, "env_variables", envVariables)
	tflog.Debug(ctx, "secrets", secrets)

	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, id),
		"input_env_variables": inputEnvVariables,
		"input_secrets":       inputSecrets,
		"env_variables":       MapToTfValue(envVariables),
		"secrets":             MapToTfValue(secrets),
		"secret_refs":         MapToTfValue(secretRefs),
		"field_refs":          FieldRefsToTfValue(ev.FieldRefs()),
		"config_map_refs":     ConfigMapRefsToTfValue(ev.ConfigMapRefs()),
		"resource_field_refs": ResourceFieldRefsToTfValue(ev.ResourceFieldRefs()),
		"file_key_refs":       FileKeyRefsToTfValue(ev.FileKeyRefs()),
	}, nil, nil
}
