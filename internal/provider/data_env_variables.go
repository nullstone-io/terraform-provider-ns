package provider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"regexp"
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

func (d *dataEnvVariables) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	inputEnvVariables := extractMapFromConfig(config, "input_env_variables")
	inputSecrets := extractMapFromConfig(config, "input_secrets")

	tflog.Debug(ctx, "input_env_variables", inputEnvVariables)
	tflog.Debug(ctx, "input_secrets", inputSecrets)

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
				tflog.Debug(ctx, fmt.Sprintf("Found and replaced secret (%s) in env variable: %s", key, k), result)
				delete(envVariables, k)
				secrets[k] = tftypes.NewValue(tftypes.String, result)
			}
		}
		// now do any replacements in the other secrets
		for k, v := range secrets {
			// we don't want to replace the secret with itself (this will prevent an infinite loop)
			if k != key {
				result := regex.ReplaceAllString(extractStringFromTfValue(v), extractStringFromTfValue(secret))
				if result != extractStringFromTfValue(v) {
					tflog.Debug(ctx, fmt.Sprintf("Found and replaced secret (%s) in secret: %s", key, k), result)
				}
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
				if result != extractStringFromTfValue(v) {
					tflog.Debug(ctx, fmt.Sprintf("Found and replaced env variable (%s) in env variable: %s", key, k), result)
				}
				envVariables[k] = tftypes.NewValue(tftypes.String, result)
			}
		}
		for k, v := range secrets {
			result := regex.ReplaceAllString(extractStringFromTfValue(v), extractStringFromTfValue(value))
			if result != extractStringFromTfValue(v) {
				tflog.Debug(ctx, fmt.Sprintf("Found and replaced env variable (%s) in secret: %s", key, k), result)
			}
			secrets[k] = tftypes.NewValue(tftypes.String, result)
		}
	}

	// calculate the unique id for this data source based on a hash of the resulting env variables and secrets
	id := d.HashFromValues(envVariables, secrets)

	tflog.Debug(ctx, "id", id)
	tflog.Debug(ctx, "env_variables", envVariables)
	tflog.Debug(ctx, "secrets", secrets)

	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, id),
		"input_env_variables": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, inputEnvVariables),
		"input_secrets":       tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, inputSecrets),
		"env_variables":       tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, envVariables),
		"secrets":             tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, secrets),
	}, nil, nil
}

func (d *dataEnvVariables) HashFromValues(envVariables, secrets map[string]tftypes.Value) string {
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
