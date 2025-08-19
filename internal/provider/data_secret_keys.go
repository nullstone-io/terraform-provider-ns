package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type dataSecretKeys struct {
	p *provider
}

func newDataSecretKeys(p *provider) (*dataSecretKeys, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataSecretKeys{p: p}, nil
}

func (*dataSecretKeys) Schema(ctx context.Context) *tfprotov5.Schema {
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
			Name:            "input_secret_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "The raw secrets before they are interpolated.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Required:        true,
			Sensitive:       true,
		},
		{
			Name:            "secret_keys",
			Type:            tftypes.Set{ElementType: tftypes.String},
			Description:     "The keys of all the secrets.",
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

func (d *dataSecretKeys) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	inputEnvVariables := TfValueToMap(config["input_env_variables"])
	inputSecretKeys := TfSetValueToStringSlice(config["input_secret_keys"])

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
	for _, key := range inputSecretKeys {
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

func (d *dataSecretKeys) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	inputEnvVariables := config["input_env_variables"]
	inputSecretKeys := config["input_secret_keys"]

	// Shuffle secret keys slice into a map so we can use Interpolate to check secret keys
	inputSecrets := map[string]string{}
	for _, v := range TfSetValueToStringSlice(inputSecretKeys) {
		inputSecrets[v] = ""
	}

	ev := NewEnvVars(TfValueToMap(inputEnvVariables), inputSecrets)
	ev.Interpolate()

	id := ev.KeysHash()
	secretKeys := ev.SecretKeys()

	tflog.Debug(ctx, "Read Secrets", map[string]interface{}{
		"id":                  id,
		"input_env_variables": inputEnvVariables,
		"input_secret_keys":   inputSecretKeys,
		"secret_keys":         secretKeys,
	})

	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, id),
		"input_env_variables": inputEnvVariables,
		"input_secret_keys":   inputSecretKeys,
		"secret_keys":         SliceToTfSet(secretKeys),
	}, nil, nil
}
