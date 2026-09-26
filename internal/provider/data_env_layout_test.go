package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataEnvLayout(t *testing.T) {
	getNsConfig, _ := mockNs(nil)
	getTfeConfig, _ := mockTfe(nil)
	factories := protoV5ProviderFactories(getNsConfig, getTfeConfig, nil)

	t.Run("classifies secret keys without values", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_layout" "this" {
  platform = "ecs"
  standard = {
    NULLSTONE_STACK = "primary"
    NULLSTONE_ENV   = "dev"
  }
  capability_env = [
    { capability = "postgres0", name = "HOST", value = "db.internal" },
  ]
  capability_secret_keys = [
    { capability = "postgres0", name = "PASSWORD" },
  ]
  capability_prefixes = {
    postgres0 = "PG_"
  }
  user_env = {
    DATABASE_URL = "postgres://app:{{ PG_PASSWORD }}@{{ PG_HOST }}/app"
    EXISTING     = "{{ secret(arn:aws:secretsmanager:us-east-1:0123456789012:secret:existing) }}"
  }
  user_secret_keys = ["API_TOKEN"]
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.ns_env_layout.this", "id"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "managed_secret_keys.#", "3"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "managed_secret_keys.*", "API_TOKEN"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "managed_secret_keys.*", "PG_PASSWORD"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "managed_secret_keys.*", "DATABASE_URL"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "unmanaged_secret_keys.#", "1"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "unmanaged_secret_keys.*", "EXISTING"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "all_secret_keys.#", "4"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.%", "7"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.NULLSTONE_STACK", "standard"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.PG_HOST", "capability"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.PG_PASSWORD", "capability"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.API_TOKEN", "user"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "capabilities.%", "2"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "capabilities.PG_PASSWORD", "postgres0"),
					),
				},
			},
		})
	})

	t.Run("plan-time validation rejects invalid keys", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_layout" "this" {
  platform = "ecs"
  user_secret_keys = ["BAD-KEY"]
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Invalid\s+environment\s+variable\s+key:\s+BAD-KEY`),
				},
			},
		})
	})
}
