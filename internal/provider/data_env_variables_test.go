package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataVariables(t *testing.T) {
	arn := "arn:aws:secretsmanager:us-east-1:0123456789012:secret:my_little_secret"

	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "input_env_variables.%", "7"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "input_secrets.%", "1"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "env_variables.%", "5"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "env_variables.FEATURE_FLAG_0115", "true"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "env_variables.IDENTIFIER", "primary.acme-api.dev"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "secrets.%", "2"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "secrets.POSTGRES_URL", "postgres://user:pass@host:port/db"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "secrets.DATABASE_URL", "postgres://user:pass@host:port/db"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "secret_refs.%", "1"),
		resource.TestCheckResourceAttr("data.ns_env_variables.this", "secret_refs.VAR_WITH_REF", arn),
	)

	t.Run("sets up attributes properly hard-coded", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_env_variables" "this" {
	input_env_variables = {
		NULLSTONE_STACK = "primary"
		NULLSTONE_BLOCK = "acme-api"
		NULLSTONE_ENV = "dev"
		FEATURE_FLAG_0115 = "true"
		DATABASE_URL = "{{POSTGRES_URL}}"
		IDENTIFIER = "{{ NULLSTONE_STACK }}.{{ NULLSTONE_BLOCK }}.{{ NULLSTONE_ENV }}"
		VAR_WITH_REF = "{{ secret(%s) }}"
	}
	input_secrets = {
		POSTGRES_URL = "postgres://user:pass@host:port/db"
	}
}
`, arn)
		getNsConfig, _ := mockNs(nil)
		getTfeConfig, _ := mockTfe(nil)

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig, nil),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks,
				},
			},
		})
	})
}
