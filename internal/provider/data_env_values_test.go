package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// testCheckAttrNotContains asserts that a state attribute exists and does not contain the given substring.
func testCheckAttrNotContains(name, key, needle string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		value, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: attribute %q not found", name, key)
		}
		if strings.Contains(value, needle) {
			return fmt.Errorf("%s: attribute %q must not contain %q", name, key, needle)
		}
		return nil
	}
}

func TestDataEnvValues(t *testing.T) {
	getNsConfig, _ := mockNs(nil)
	getTfeConfig, _ := mockTfe(nil)
	factories := protoV5ProviderFactories(getNsConfig, getTfeConfig, nil)

	t.Run("ecs: layers, prefixes, promotion, platform_data", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_values" "this" {
  platform = "aws_ecs"
  standard = {
    NULLSTONE_STACK = "primary"
    NULLSTONE_ENV   = "dev"
  }
  cloud = {
    AWS_REGION = "us-east-1"
  }
  capability_env = [
    { capability = "postgres0", name = "HOST", value = "db.internal" },
  ]
  capability_secrets = [
    { capability = "postgres0", name = "PASSWORD", value = "hunter2" },
  ]
  capability_prefixes = {
    postgres0 = "PG_"
  }
  user_env = {
    IDENTIFIER   = "{{ NULLSTONE_STACK }}.{{ NULLSTONE_ENV }}"
    DATABASE_URL = "postgres://app:{{ PG_PASSWORD }}@{{ PG_HOST }}/app"
    AWS_REGION   = "us-west-2"
  }
  user_secrets = {
    API_TOKEN = "tok-123"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.ns_env_values.this", "id"),
						// plain
						resource.TestCheckResourceAttr("data.ns_env_values.this", "env_variables.%", "5"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "env_variables.NULLSTONE_ENV", "dev"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "env_variables.NULLSTONE_STACK", "primary"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "env_variables.IDENTIFIER", "primary.dev"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "env_variables.PG_HOST", "db.internal"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "env_variables.AWS_REGION", "us-west-2"),
						resource.TestCheckNoResourceAttr("data.ns_env_values.this", "env_variables.DATABASE_URL"),
						// managed secrets
						resource.TestCheckResourceAttr("data.ns_env_values.this", "secrets.%", "3"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "secrets.DATABASE_URL", "postgres://app:hunter2@db.internal/app"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "secrets.PG_PASSWORD", "hunter2"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "secrets.API_TOKEN", "tok-123"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "managed_secret_keys.#", "3"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_values.this", "managed_secret_keys.*", "DATABASE_URL"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_values.this", "managed_secret_keys.*", "PG_PASSWORD"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_values.this", "managed_secret_keys.*", "API_TOKEN"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "unmanaged_secret_keys.#", "0"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "all_secret_keys.#", "3"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "unmanaged_secret_refs.%", "0"),
						// sources
						resource.TestCheckResourceAttr("data.ns_env_values.this", "sources.%", "8"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "sources.NULLSTONE_STACK", "standard"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "sources.AWS_REGION", "user"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "sources.PG_HOST", "capability"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "sources.PG_PASSWORD", "capability"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "sources.DATABASE_URL", "user"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "capabilities.%", "2"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "capabilities.PG_HOST", "postgres0"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "capabilities.PG_PASSWORD", "postgres0"),
						// k8s refs empty
						resource.TestCheckResourceAttr("data.ns_env_values.this", "field_refs.%", "0"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "config_map_refs.%", "0"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "resource_field_refs.%", "0"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "file_key_refs.%", "0"),
						// platform_data
						resource.TestMatchResourceAttr("data.ns_env_values.this", "platform_data", regexp.MustCompile(`"platform":"aws_ecs"`)),
						resource.TestMatchResourceAttr("data.ns_env_values.this", "platform_data", regexp.MustCompile(`"source":"capability"`)),
						resource.TestMatchResourceAttr("data.ns_env_values.this", "platform_data", regexp.MustCompile(`"PG_PASSWORD":\{"sensitive":true,"source":"capability","capability":"postgres0"\}`)),
						resource.TestMatchResourceAttr("data.ns_env_values.this", "platform_data", regexp.MustCompile(`"IDENTIFIER":\{"template":"\{\{ NULLSTONE_STACK \}\}.\{\{ NULLSTONE_ENV \}\}","value":"primary.dev","source":"user"\}`)),
						testCheckAttrNotContains("data.ns_env_values.this", "platform_data", "hunter2"),
						testCheckAttrNotContains("data.ns_env_values.this", "platform_data", "tok-123"),
					),
				},
			},
		})
	})

	t.Run("k8s: field refs and unmanaged secret refs", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_values" "this" {
  platform = "gcp_gke"
  standard = {
    NULLSTONE_ENV = "dev"
  }
  user_env = {
    POD_IP     = "{{ k8s.field(v1, status.podIP) }}"
    EXISTING   = "{{ secret(arn:aws:secretsmanager:us-east-1:0123456789012:secret:existing) }}"
  }
  user_secrets = {
    API_TOKEN = "tok-123"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.ns_env_values.this", "env_variables.%", "1"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "field_refs.%", "1"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "field_refs.POD_IP.api_version", "v1"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "field_refs.POD_IP.field_path", "status.podIP"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "unmanaged_secret_refs.%", "1"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "unmanaged_secret_refs.EXISTING", "arn:aws:secretsmanager:us-east-1:0123456789012:secret:existing"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "unmanaged_secret_keys.#", "1"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_values.this", "unmanaged_secret_keys.*", "EXISTING"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "managed_secret_keys.#", "1"),
						resource.TestCheckTypeSetElemAttr("data.ns_env_values.this", "managed_secret_keys.*", "API_TOKEN"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "all_secret_keys.#", "2"),
						resource.TestCheckResourceAttr("data.ns_env_values.this", "secrets.%", "1"),
						resource.TestCheckNoResourceAttr("data.ns_env_values.this", "secrets.EXISTING"),
						resource.TestMatchResourceAttr("data.ns_env_values.this", "platform_data", regexp.MustCompile(`"POD_IP":\{"template":"\{\{ k8s.field\(v1, status.podIP\) \}\}","ref":\{"type":"k8s_field","api_version":"v1","field_path":"status.podIP"\},"source":"user"\}`)),
						resource.TestMatchResourceAttr("data.ns_env_values.this", "platform_data", regexp.MustCompile(`"EXISTING":\{"template":"[^"]+","sensitive":true,"ref":\{"type":"secret","id":"arn:aws:secretsmanager:us-east-1:0123456789012:secret:existing"\},"source":"user"\}`)),
					),
				},
			},
		})
	})

	t.Run("ecs rejects k8s templates", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_values" "this" {
  platform = "aws_ecs"
  user_env = {
    POD_IP = "{{ k8s.field(v1, status.podIP) }}"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Unsupported\s+template\s+for\s+platform\s+aws_ecs:\s+POD_IP`),
				},
			},
		})
	})

	t.Run("unknown platform errors", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_values" "this" {
  platform = "mainframe"
  user_env = {
    A = "b"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Unknown\s+platform:\s+mainframe`),
				},
			},
		})
	})

	t.Run("duplicate capability key errors", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_values" "this" {
  platform = "aws_ecs"
  capability_env = [
    { capability = "cap_a", name = "HOST", value = "a" },
    { capability = "cap_b", name = "HOST", value = "b" },
  ]
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Duplicate\s+capability\s+environment\s+variable:\s+HOST`),
				},
			},
		})
	})
}

// Generated capabilities.tf items carry both the legacy `cap_tf_id` and `capability`;
// Terraform must drop the extra attribute when converting to the provider's object type
// so modules can pass `local.capabilities.env` verbatim.
func TestDataEnvValues_ExtraCapabilityAttributesTolerated(t *testing.T) {
	getNsConfig, _ := mockNs(nil)
	getTfeConfig, _ := mockTfe(nil)
	factories := protoV5ProviderFactories(getNsConfig, getTfeConfig, nil)

	config := `
provider "ns" {
  organization = "org0"
}
locals {
  capabilities = {
    env = [
      { cap_tf_id = "postgres0", capability = "postgres0", name = "HOST", value = "db.internal" },
    ]
    secrets = [
      { cap_tf_id = "postgres0", capability = "postgres0", name = "PASSWORD", value = "hunter2" },
    ]
  }
  cap_prefixes = { postgres0 = "PG_" }
}
data "ns_env_values" "this" {
  platform            = "aws_ecs"
  capability_env      = local.capabilities.env
  capability_secrets  = local.capabilities.secrets
  capability_prefixes = local.cap_prefixes
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ns_env_values.this", "env_variables.PG_HOST", "db.internal"),
					resource.TestCheckResourceAttr("data.ns_env_values.this", "capabilities.PG_HOST", "postgres0"),
					resource.TestCheckResourceAttr("data.ns_env_values.this", "capabilities.PG_PASSWORD", "postgres0"),
					resource.TestCheckResourceAttr("data.ns_env_values.this", "managed_secret_keys.#", "1"),
				),
			},
		},
	})
}
