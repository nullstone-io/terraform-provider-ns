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

	ev := NewEnvVars(TfValueToMap(inputEnvVariables), TfValueToMap(inputSecrets))
	ev.Interpolate()

	// calculate the unique id for this data source based on a hash of the resulting env variables and secrets
	id := ev.Hash()
	envVariables := ev.EnvVars()
	secrets := ev.Secrets()
	secretRefs := ev.SecretRefs()

	tflog.Debug(ctx, "Read EnvVariables", map[string]interface{}{
		"id":                  id,
		"input_env_variables": inputEnvVariables,
		"input_secrets":       inputSecrets,
		"env_variables":       envVariables,
		"secrets":             secrets,
		"secret_refs":         secretRefs,
	})

	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, id),
		"input_env_variables": inputEnvVariables,
		"input_secrets":       inputSecrets,
		"env_variables":       MapToTfValue(envVariables),
		"secrets":             MapToTfValue(secrets),
		"secret_refs":         MapToTfValue(secretRefs),
	}, nil, nil
}
