package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestSecretKeys(t *testing.T) {
	arn := "arn:aws:secretsmanager:us-east-1:0123456789012:secret:my_little_secret"

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

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "input_env_variables.%", "8"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "input_secret_keys.#", "2"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.#", "4"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.0", "DATABASE_URL"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.1", "DUPLICATE_TEST"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.2", "POSTGRES_URL"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.3", "SECRET_KEY_BASE"),
		)

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

	t.Run("real scenario 1", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
variable "env_vars" {
	type    = map(string)
	default = {
		"OTEL_EXPORTER_OTLP_ENDPOINT" = "https://otlp.uptrace.dev"
		"OTEL_METRICS_EXEMPLAR_FILTER" = "always_on"
		"OTEL_METRICS_EXPORTER" = "otlp"
		"OTEL_METRIC_EXPORT_INTERVAL" = "15000"
		"OTEL_TRACES_EXPORTER" = "otlp"
		"OTEL_TRACES_SAMPLER" = "always_on"
		"POSTGRES_SLOW_QUERY" = "1m"
		"SERVICE_DOMAIN" = "{{ NULLSTONE_ENV }}"
	}
}
variable "secrets" {
	type      = map(string)
    sensitive = true
	default   = {
		"OTEL_EXPORTER_OTLP_HEADERS" = "uptrace-dsn=https://XgHrvhymCpGP20K7vUNp2A@api.uptrace.dev?grpc=4317"
  		"ROLLBAR_ACCESS_TOKEN" = "39ec403290c94fc6b81d17a308a2909d"
    }
}
locals {
	cap_env_vars = {
		"POSTGRES_DB" = "nullfire"
		"POSTGRES_HOST" = "raspberry-lizard-iznie.cs4cyqrf5rxq.us-east-1.rds.amazonaws.com"
		"POSTGRES_USER" = "nullfire-vipgg"
		"REDIS_HOST" = "master.sangria-snail-ttfuq.ejk9gq.use1.cache.amazonaws.com"
		"REDIS_PORT" = "6379"
		"TEMPORAL_HOSTPORT" = "core-prod-bvhai.i0cev.tmprl.cloud:7233"
		"TEMPORAL_NAMESPACE" = "core-prod-bvhai.i0cev"
		"TEMPORAL_TLS_CERT" = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVOVENDQXAyZ0F3SUJBZ0lSQVBqYTRWMkI1bmtwSTVuc25JN0pEbk13RFFZSktvWklodmNOQVFFTEJRQXcKS3pFU01CQUdBMVVFQ2hNSlRuVnNiSE4wYjI1bE1SVXdFd1lEVlFRREV3eHVkV3hzYzNSdmJtVXVhVzh3SGhjTgpNalF3TkRJNU1qTTBOelF6V2hjTk16UXdOREkzTWpNME56UXpXakFyTVJJd0VBWURWUVFLRXdsT2RXeHNjM1J2CmJtVXhGVEFUQmdOVkJBTVRERzUxYkd4emRHOXVaUzVwYnpDQ0FhSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnR1AKQURDQ0FZb0NnZ0dCQU1DYkt4QVY5WndmT1BURExacE9TUjh1MzlsWGE1a29kZTdoRjJSNEJ4Mng5cS8yR2RNVQpUVEZMVjYvcFo2eGhoOUl0N1ZEUjJRQUUzQXgyNnhPWnZUck9Jd0RaWEZVbXdQMG9jcEZXVENEVjJhb0Z2cUk3Cno5NmR2SGsyTUlHb0RWWVc5ejBoS0NhOEdrSnpLbG55dWM3ekZONUhoQzJzRjlhdzREMVlSTGcwK0t3NWdsZUwKamFLSFpySVlRMFA3Z040QVJFeDNRd2QrQ0d4VWtXQysyWnRmN01PK3dSeTBESncxcW0wRVVBNTZNWlM1SnpRQwpuZnFHTXFGNmNrYlRuZEFFQ0ZKRERka2YwRGpyUVkrallETnFKVFlWSCt4eDMycWpyTDMvYkxSdGhwK0ZPUlJQCnR3bUFzWHNZU1FpWjB1N2s3WUZud1ZPUWRGNHB0SHFRMzByZVlXK0tTV2tlOVpaSEJDclExeStqV2hwRnh3cTAKTTZrS3BFajFSZDA4aFVCTjdGYlJ6TTlkdzU1OHpOUGZDYklkdnYzYnpKdzBSb1RZaWVwb0F4NGFkWlJkcWx4bApQUFNPZHpJUi9CbUhZTkZ6UnVhSllIdUtGL1NtUmJWVWNTT0k0cG9pR3ZoNHF2T1pVNHFGRmlCaXpndWhxaU1zCk5FTVc2em5ZRGxnNFB3SURBUUFCbzFRd1VqQU9CZ05WSFE4QkFmOEVCQU1DQmFBd0hRWURWUjBsQkJZd0ZBWUkKS3dZQkJRVUhBd0VHQ0NzR0FRVUZCd01DTUF3R0ExVWRFd0VCL3dRQ01BQXdFd1lEVlIwUkJBd3dDb0lJYm5WcwpiR1pwY21Vd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dHQkFESVdLeFd0WG84S3hDT1liZDMwZGJMSTcvbTRtOU4zClMyTGYraXU0ZzJKTndqcXhmODRIemVUK3ZGRG14MExWbHI1Y2pPbXU3bW9jdXo2bFZUNmNGc0FXaXpuQVcwZ00KQTFMRnZzSkY4RGxsWHI4ZE50NnIyd0poOVoycFhCdWFkOWxWSU0zUTJLa2xKbW9tY3l6T1QxRmFPYzk3RnQ0ZgpMWGtaWkpmZE50YmhNMHJwY0FXTUhXbWhRaTFOcUdIYUtSQzNwMk5VclpyZ0dway9MV0NrSHNBdytHZ3kvZVM2CkJrWEhRSGlrYjFnZFFOVDVNdGlmOXZidUpKSWt4OTRUUHBrZWZmalduWWxsekVRVEtmV2FobzA2VlF1enhQVDgKaEdCb2ZnaDM2SElGOUhFWGNQQm9XT2RNMTdJV0Q0RS8yZEZmQ2hOTlhQd2hwV0xLVjR3Lzk4aERldUh2Sk5UUgpSMnY5Z2xzSUlvQmpPWDlLR212T3BadFRIOGtBeTQybXh2aEVSdFZnNVdKUlcvbG1CdU14a0d6RWJ1UUpjTVRJCjVqRlhLNndZQjBjTkVlcWxyVnhPQ0ZvN3pBV29TdzdIOGFySXhLc29Gbnd2bHNSU1BialdyaHA3bXZhUmdnUWUKK29nMnhjeEV1NENUOE9yZzdLbFhQRTRqa0hua010eFl5Zz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
	}
	cap_secrets = {
		"POSTGRES_PASSWORD" = sensitive("something")
		"POSTGRES_URL" = sensitive("something")
		"REDIS_AUTH_TOKEN" = sensitive("something")
		"REDIS_URL" = sensitive("something")
		"TEMPORAL_TLS_KEY" = sensitive("something")
    }
}
locals {
	standard_env_vars = tomap({
		NULLSTONE_STACK         = "core"
		NULLSTONE_APP           = "api1"
		NULLSTONE_ENV           = "dev"
		NULLSTONE_VERSION       = "0.2.3"
	})
	
	input_env_vars    = merge(local.standard_env_vars, local.cap_env_vars, var.env_vars)
	input_secrets     = merge(local.cap_secrets, var.secrets)
	input_secret_keys = nonsensitive(concat(keys(local.cap_secrets), keys(var.secrets)))
}
data "ns_env_variables" "this" {
  input_env_variables = local.input_env_vars
  input_secrets       = local.input_secrets
}
data "ns_secret_keys" "this" {
  //input_env_variables = local.input_env_vars
  //input_secret_keys   = local.input_secret_keys
  input_env_variables = local.input_env_vars
  input_secret_keys   = nonsensitive(keys(local.input_secrets))
}
`)
		getNsConfig, _ := mockNs(nil)
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "input_env_variables.%", "20"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "input_secret_keys.#", "7"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.#", "7"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.0", "OTEL_EXPORTER_OTLP_HEADERS"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.1", "POSTGRES_PASSWORD"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.2", "POSTGRES_URL"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.3", "REDIS_AUTH_TOKEN"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.4", "REDIS_URL"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.5", "ROLLBAR_ACCESS_TOKEN"),
			resource.TestCheckResourceAttr("data.ns_secret_keys.this", "secret_keys.6", "TEMPORAL_TLS_KEY"),
		)

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
