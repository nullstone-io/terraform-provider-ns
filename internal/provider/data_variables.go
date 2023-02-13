package provider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"regexp"
)

type dataVariables struct {
	p *provider
}

func newDataVariables(p *provider) (*dataVariables, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataVariables{p: p}, nil
}

func (*dataVariables) Schema(ctx context.Context) *tfprotov5.Schema {
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
			Name:            "env_variable_keys",
			Type:            tftypes.List{ElementType: tftypes.String},
			Description:     "The keys of all the environment variables.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "env_variables",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Description:     "The processed environment variables after they are interpolated.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "secret_keys",
			Type:            tftypes.List{ElementType: tftypes.String},
			Description:     "The keys of all the secrets.",
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

func (d *dataVariables) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	inputEnvVariables := extractMapFromConfig(config, "input_env_variables")
	inputSecrets := extractMapFromConfig(config, "input_secrets")

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

func (d *dataVariables) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	inputEnvVariables := extractMapFromConfig(config, "input_env_variables")
	inputSecrets := extractMapFromConfig(config, "input_secrets")

	// make sure we copy these so our changes below don't affect the original values
	envVariables := copyMap(inputEnvVariables)
	secrets := copyMap(inputSecrets)

	regexPattern := "{{\\s?%s\\s?}}"

	// we are going to first loop through all the input secrets
	//   find and replace this secret in all the rest of the env variables and secrets
	for key, secret := range secrets {
		regex := regexp.MustCompile(fmt.Sprintf(regexPattern, key))
		// first try and replace in the env variables
		for k, v := range envVariables {
			result := regex.ReplaceAllString(extractStringFromTfValue(v), extractStringFromTfValue(secret))
			// if a match was found and replaced, this env variable is now a secret
			if result != extractStringFromTfValue(v) {
				delete(envVariables, k)
				secrets[k] = tftypes.NewValue(tftypes.String, result)
			}
		}
		// now do any replacements in the other secrets
		for k, v := range secrets {
			// we don't want to replace the secret with itself (this will prevent an infinite loop)
			if k != key {
				result := regex.ReplaceAllString(extractStringFromTfValue(v), extractStringFromTfValue(secret))
				secrets[k] = tftypes.NewValue(tftypes.String, result)
			}
		}
	}

	// now we will loop through all the env variables
	//   find and replace this env variable in all the rest of the env variables and secrets
	for key, value := range envVariables {
		regex := regexp.MustCompile(fmt.Sprintf(regexPattern, key))
		for k, v := range envVariables {
			// we don't want to replace the env variable with itself (this will prevent an infinite loop)
			if k != key {
				result := regex.ReplaceAllString(extractStringFromTfValue(v), extractStringFromTfValue(value))
				envVariables[k] = tftypes.NewValue(tftypes.String, result)
			}
		}
		for k, v := range secrets {
			result := regex.ReplaceAllString(extractStringFromTfValue(v), extractStringFromTfValue(value))
			secrets[k] = tftypes.NewValue(tftypes.String, result)
		}
	}

	// extract the keys from the maps
	envVariableKeys := make([]tftypes.Value, 0, len(envVariables))
	for k, _ := range envVariables {
		envVariableKeys = append(envVariableKeys, tftypes.NewValue(tftypes.String, k))
	}
	secretKeys := make([]tftypes.Value, 0, len(secrets))
	for k, _ := range secrets {
		secretKeys = append(secretKeys, tftypes.NewValue(tftypes.String, k))
	}

	// calculate the unique id for this data source based on a hash of the resulting env variables and secrets
	id := d.HashFromValues(envVariables, secrets)

	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, id),
		"input_env_variables": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, inputEnvVariables),
		"input_secrets":       tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, inputSecrets),
		"env_variable_keys":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, envVariableKeys),
		"env_variables":       tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, envVariables),
		"secret_keys":         tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, secretKeys),
		"secrets":             tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, secrets),
	}, nil, nil
}

func (d *dataVariables) HashFromValues(envVariables, secrets map[string]tftypes.Value) string {
	hashString := ""
	for k, v := range envVariables {
		hashString += fmt.Sprintf("%s=%s;", k, extractStringFromTfValue(v))
	}
	for k, v := range secrets {
		hashString += fmt.Sprintf("%s=%s;", k, extractStringFromTfValue(v))
	}

	sum := sha256.Sum256([]byte(hashString))
	return fmt.Sprintf("%x", sum)
}
