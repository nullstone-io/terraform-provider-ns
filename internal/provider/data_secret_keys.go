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
	inputEnvVariables := extractMapFromConfig(config, "input_env_variables")
	inputSecretKeys := extractSetFromConfig(config, "input_secret_keys")

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
		if !validEnvVariableKey(extractStringFromTfValue(key)) {
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
	inputEnvVariables := extractMapFromConfig(config, "input_env_variables")
	inputSecretKeys := extractSetFromConfig(config, "input_secret_keys")

	tflog.Debug(ctx, "input_env_variables", inputEnvVariables)
	tflog.Debug(ctx, "input_secret_keys", inputSecretKeys)

	// make sure we copy these so our changes below don't affect the original values
	secretKeys := copySet(inputSecretKeys)

	regexPattern := "{{\\s?%s\\s?}}"

	// loop through and determine if any of the environment variables contain interpolation using any of the secret keys
	//   if they do, add their keys to the final set of secret keys
	added := map[string]bool{}
	for _, secretKey := range inputSecretKeys {
		regex := regexp.MustCompile(fmt.Sprintf(regexPattern, extractStringFromTfValue(secretKey)))
		// first try and replace in the env variables
		for k, v := range inputEnvVariables {
			// don't add the key more than once, the "added" map keeps track of whether it has already been added
			if added[k] {
				continue
			}
			if found := regex.MatchString(extractStringFromTfValue(v)); found {
				secretKeys = append(secretKeys, tftypes.NewValue(tftypes.String, k))
				added[k] = true
			}
		}
	}

	// calculate the unique id for this data source based on a hash of the resulting env variables and secrets
	id := d.HashFromValues(secretKeys)

	tflog.Debug(ctx, "id", id)
	tflog.Debug(ctx, "secret_keys", secretKeys)

	return map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, id),
		"input_env_variables": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, inputEnvVariables),
		"input_secret_keys":   tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, inputSecretKeys),
		"secret_keys":         tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, secretKeys),
	}, nil, nil
}

func (d *dataSecretKeys) HashFromValues(secretKeys []tftypes.Value) string {
	hashString := ""
	for _, v := range secretKeys {
		hashString += fmt.Sprintf("%s=%s;", v, extractStringFromTfValue(v))
	}

	sum := sha256.Sum256([]byte(hashString))
	return fmt.Sprintf("%x", sum)
}
