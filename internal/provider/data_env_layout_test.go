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

	t.Run("classifies secret keys from key-only layers and user_env templates", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_layout" "this" {
  platform      = "aws_ecs"
  standard_keys = ["NULLSTONE_STACK", "NULLSTONE_ENV"]
  cloud_keys    = ["AWS_REGION"]
  otel_keys     = ["OTEL_SERVICE_NAME"]
  capability_env_keys = [
    { capability = "postgres0", name = "HOST" },
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
    AWS_REGION   = "us-west-2"
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
						// {{ PG_PASSWORD }} promotes DATABASE_URL to a managed secret
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "managed_secret_keys.#", "3"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "managed_secret_keys.*", "API_TOKEN"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "managed_secret_keys.*", "PG_PASSWORD"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "managed_secret_keys.*", "DATABASE_URL"),
						// {{ secret(arn) }} is unmanaged
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "unmanaged_secret_keys.#", "1"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "unmanaged_secret_keys.*", "EXISTING"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "all_secret_keys.#", "4"),
						// sources: key-only layers are still attributed
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.%", "9"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.NULLSTONE_STACK", "standard"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.AWS_REGION", "user"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.OTEL_SERVICE_NAME", "otel"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.PG_HOST", "capability"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.PG_PASSWORD", "capability"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.API_TOKEN", "user"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "capabilities.%", "2"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "capabilities.PG_HOST", "postgres0"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "capabilities.PG_PASSWORD", "postgres0"),
					),
				},
			},
		})
	})

	t.Run("secret wins over user_env at the same key", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_layout" "this" {
  platform = "aws_ecs"
  capability_secret_keys = [
    { capability = "postgres0", name = "PASSWORD" },
  ]
  capability_prefixes = {
    postgres0 = "PG_"
  }
  user_env = {
    PG_PASSWORD = "not-a-secret"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "managed_secret_keys.#", "1"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_layout.this", "managed_secret_keys.*", "PG_PASSWORD"),
						resource.TestCheckResourceAttr("data.ns_env_layout.this", "sources.PG_PASSWORD", "capability"),
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
  platform = "aws_ecs"
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

	t.Run("plan-time validation rejects invalid standard_keys", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_layout" "this" {
  platform      = "aws_ecs"
  standard_keys = ["1BAD"]
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Invalid\s+environment\s+variable\s+key:\s+1BAD`),
				},
			},
		})
	})

	t.Run("capability missing from capability_prefixes errors", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_layout" "this" {
  platform = "aws_ecs"
  capability_env_keys = [
    { capability = "postgres0", name = "HOST" },
  ]
  capability_prefixes = {
    redis0 = "REDIS_"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Unknown\s+capability:\s+postgres0`),
				},
			},
		})
	})

	t.Run("empty capability name errors", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_layout" "this" {
  platform = "aws_ecs"
  capability_secret_keys = [
    { capability = "", name = "PASSWORD" },
  ]
  capability_prefixes = {}
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Capability\s+name\s+is\s+required`),
				},
			},
		})
	})
}
