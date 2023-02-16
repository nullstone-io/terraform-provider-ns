package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestSecretKeys(t *testing.T) {
	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "input_env_variables.%", "6"),
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "input_secret_keys.#", "1"),
		resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.#", "2"),
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
	}
	input_secret_keys = [
		"POSTGRES_URL"
	]
}
`)
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
