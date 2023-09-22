package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestSecretKeys(t *testing.T) {
	arn := "arn:aws:secretsmanager:us-east-1:522657839841:secret:scarlet-eagle-kvoty/conn_url-lPd8oL"

	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "input_env_variables.%", "8"),
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "input_secret_keys.#", "2"),
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.#", "4"),
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.0", "DATABASE_URL"),
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.1", "DUPLICATE_TEST"),
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.2", "POSTGRES_URL"),
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.3", "SECRET_KEY_BASE"),
	)

	t.Run("sets up attributes properly hard-coded", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_secret_keys" "this" {
	input_env_variables = {
		NULLSTONE_STACK = "primary"
		NULLSTONE_BLOCK = "acme-api"
		NULLSTONE_ENV = "dev"
		FEATURE_FLAG_0115 = "true"
		DATABASE_URL = "{{POSTGRES_URL}}"
		IDENTIFIER = "{{ NULLSTONE_STACK }}.{{ NULLSTONE_BLOCK }}.{{ NULLSTONE_ENV }}"
		DUPLICATE_TEST = "{{ SECRET_KEY_BASE }}{{ POSTGRES_URL }}"
		VAR_WITH_REF = "{{ secret(%s) }}"
	}
	input_secret_keys = [
		"POSTGRES_URL",
		"SECRET_KEY_BASE"
	]
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
