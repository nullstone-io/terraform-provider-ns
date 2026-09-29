package provider

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nullstone-io/module/platformdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// envValuesFixture is the ns_env_values block shared by the ns_env_platform_data tests.
func envValuesFixture(platform string) string {
	return `
provider "ns" {
  organization = "org0"
}
data "ns_env_values" "this" {
  platform = "` + platform + `"
  standard = {
    NULLSTONE_ENV = "dev"
  }
  capability_secrets = [
    { capability = "postgres0", name = "PASSWORD", value = "hunter2" },
  ]
  capability_prefixes = {
    postgres0 = "PG_"
  }
  user_env = {
    EXISTING = "{{ secret(arn:aws:secretsmanager:us-east-1:0123456789012:secret:existing) }}"
  }
  user_secrets = {
    API_TOKEN = "tok-123"
  }
}
`
}

func TestDataEnvPlatformData(t *testing.T) {
	getNsConfig, _ := mockNs(nil)
	getTfeConfig, _ := mockTfe(nil)
	factories := protoV5ProviderFactories(getNsConfig, getTfeConfig, nil)

	t.Run("aws: secret ids attach to every managed secret", func(t *testing.T) {
		config := envValuesFixture("aws_ecs") + `
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  secret_ids = {
    PG_PASSWORD = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:pg"
    API_TOKEN   = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:tok"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.ns_env_platform_data.this", "id"),
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"platform":"aws_ecs"`)),
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"PG_PASSWORD":\{"sensitive":true,"ref":\{"type":"secret","id":"arn:aws:secretsmanager:us-east-1:0123456789012:secret:pg"\},"source":"capability","capability":"postgres0"\}`)),
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"API_TOKEN":\{"sensitive":true,"ref":\{"type":"secret","id":"arn:aws:secretsmanager:us-east-1:0123456789012:secret:tok"\},"source":"user"\}`)),
						// unmanaged secret and plain values pass through untouched
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"EXISTING":\{"template":"[^"]+","sensitive":true,"ref":\{"type":"secret","id":"arn:aws:secretsmanager:us-east-1:0123456789012:secret:existing"\},"source":"user"\}`)),
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"NULLSTONE_ENV":\{"template":"dev","value":"dev","source":"standard"\}`)),
						testCheckAttrNotContains("data.ns_env_platform_data.this", "platform_data", "hunter2"),
						testCheckAttrNotContains("data.ns_env_platform_data.this", "platform_data", "tok-123"),
					),
				},
			},
		})
	})

	t.Run("gke: k8s secret refs attach on a kubernetes platform", func(t *testing.T) {
		config := envValuesFixture("gcp_gke") + `
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  k8s_secret_refs = {
    PG_PASSWORD = { name = "app-secrets", key = "PG_PASSWORD" }
    API_TOKEN   = { name = "app-secrets", key = "API_TOKEN" }
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"platform":"gcp_gke"`)),
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"PG_PASSWORD":\{"sensitive":true,"ref":\{"type":"k8s_secret_key","name":"app-secrets","key":"PG_PASSWORD"\},"source":"capability","capability":"postgres0"\}`)),
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"API_TOKEN":\{"sensitive":true,"ref":\{"type":"k8s_secret_key","name":"app-secrets","key":"API_TOKEN"\},"source":"user"\}`)),
					),
				},
			},
		})
	})

	t.Run("mixed: secret_ids and k8s_secret_refs together", func(t *testing.T) {
		config := envValuesFixture("gcp_gke") + `
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  secret_ids = {
    PG_PASSWORD = "projects/p/secrets/pg"
  }
  k8s_secret_refs = {
    API_TOKEN = { name = "app-secrets", key = "API_TOKEN" }
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"PG_PASSWORD":\{"sensitive":true,"ref":\{"type":"secret","id":"projects/p/secrets/pg"\}`)),
						resource.TestMatchResourceAttr("data.ns_env_platform_data.this", "platform_data", regexp.MustCompile(`"API_TOKEN":\{"sensitive":true,"ref":\{"type":"k8s_secret_key","name":"app-secrets","key":"API_TOKEN"\}`)),
					),
				},
			},
		})
	})

	t.Run("missing coverage errors", func(t *testing.T) {
		config := envValuesFixture("aws_ecs") + `
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  secret_ids = {
    PG_PASSWORD = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:pg"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Managed\s+secrets\s+without\s+an\s+id:\s+API_TOKEN`),
				},
			},
		})
	})

	t.Run("unknown key errors", func(t *testing.T) {
		config := envValuesFixture("aws_ecs") + `
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  secret_ids = {
    PG_PASSWORD = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:pg"
    API_TOKEN   = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:tok"
    NOPE        = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:nope"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Invalid\s+secret\s+id\s+key:\s+NOPE`),
				},
			},
		})
	})

	t.Run("non-managed key errors", func(t *testing.T) {
		config := envValuesFixture("aws_ecs") + `
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  secret_ids = {
    PG_PASSWORD   = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:pg"
    API_TOKEN     = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:tok"
    NULLSTONE_ENV = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:env"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Invalid\s+secret\s+id\s+key:\s+NULLSTONE_ENV`),
				},
			},
		})
	})

	t.Run("k8s secret refs on aws_ecs error", func(t *testing.T) {
		config := envValuesFixture("aws_ecs") + `
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  secret_ids = {
    API_TOKEN = "arn:aws:secretsmanager:us-east-1:0123456789012:secret:tok"
  }
  k8s_secret_refs = {
    PG_PASSWORD = { name = "app-secrets", key = "PG_PASSWORD" }
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Unsupported\s+secret\s+ref\s+for\s+platform\s+aws_ecs:\s+PG_PASSWORD`),
				},
			},
		})
	})

	t.Run("key in both maps errors", func(t *testing.T) {
		config := envValuesFixture("gcp_gke") + `
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  secret_ids = {
    PG_PASSWORD = "projects/p/secrets/pg"
    API_TOKEN   = "projects/p/secrets/tok"
  }
  k8s_secret_refs = {
    PG_PASSWORD = { name = "app-secrets", key = "PG_PASSWORD" }
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Secret\s+id\s+declared\s+twice:\s+PG_PASSWORD`),
				},
			},
		})
	})

	t.Run("invalid values errors at plan", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_env_platform_data" "this" {
  values = jsonencode({ nope = true })
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Invalid\s+env\s+platform\s+data`),
				},
			},
		})
	})
}

func TestDataEnvPlatformData_Validate(t *testing.T) {
	ds := &dataEnvPlatformData{p: &provider{}}
	nullIds := tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil)
	nullRefs := tftypes.NewValue(tftypes.Map{ElementType: k8sSecretRefObjectType}, nil)

	t.Run("skips validation when values is unknown", func(t *testing.T) {
		config := map[string]tftypes.Value{
			"values": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"secret_ids": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"NOPE": tftypes.NewValue(tftypes.String, "arn"),
			}),
			"k8s_secret_refs": nullRefs,
		}
		diags, err := ds.Validate(context.Background(), config)
		require.NoError(t, err)
		assert.Empty(t, diags)
	})

	t.Run("parses values but skips coverage when secret_ids is unknown", func(t *testing.T) {
		config := map[string]tftypes.Value{
			"values":          tftypes.NewValue(tftypes.String, `{"platform":"aws_ecs","variables":{"API_TOKEN":{"sensitive":true}}}`),
			"secret_ids":      tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
			"k8s_secret_refs": nullRefs,
		}
		diags, err := ds.Validate(context.Background(), config)
		require.NoError(t, err)
		assert.Empty(t, diags)
	})

	t.Run("reports invalid values", func(t *testing.T) {
		config := map[string]tftypes.Value{
			"values":          tftypes.NewValue(tftypes.String, `{}`),
			"secret_ids":      tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
			"k8s_secret_refs": nullRefs,
		}
		diags, err := ds.Validate(context.Background(), config)
		require.NoError(t, err)
		require.Len(t, diags, 1)
		assert.Equal(t, "Invalid env platform data", diags[0].Summary)
	})

	t.Run("reports missing coverage when everything is known", func(t *testing.T) {
		config := map[string]tftypes.Value{
			"values":          tftypes.NewValue(tftypes.String, `{"platform":"aws_ecs","variables":{"API_TOKEN":{"sensitive":true},"PG_PASSWORD":{"sensitive":true}}}`),
			"secret_ids":      nullIds,
			"k8s_secret_refs": nullRefs,
		}
		diags, err := ds.Validate(context.Background(), config)
		require.NoError(t, err)
		require.Len(t, diags, 1)
		assert.Equal(t, "Managed secrets without an id: API_TOKEN, PG_PASSWORD", diags[0].Summary)
	})
}

func TestCompleteEnvPlatformData(t *testing.T) {
	values := func(platform string) string {
		raw, err := json.Marshal(platformdata.EnvV1{Platform: platform, Variables: map[string]platformdata.EnvV1Variable{
			"PLAIN":     {Value: "x", Source: platformdata.SourceStandard},
			"MANAGED":   {Sensitive: true, Source: platformdata.SourceUser},
			"UNMANAGED": {Sensitive: true, Ref: &platformdata.EnvV1Ref{Type: platformdata.RefTypeSecret, Id: "existing"}, Source: platformdata.SourceUser},
		}})
		require.NoError(t, err)
		return string(raw)
	}

	t.Run("unknown platform in record still accepts k8s refs", func(t *testing.T) {
		// Platform is optional in env v1 (legacy adapter records omit it); only a known non-k8s platform rejects k8s refs.
		record, diags := completeEnvPlatformData(envPlatformDataInput{
			Values:        values(""),
			K8sSecretRefs: map[string]k8sSecretRef{"MANAGED": {Name: "s", Key: "MANAGED"}},
		})
		require.Empty(t, diags)
		assert.Equal(t, &platformdata.EnvV1Ref{Type: platformdata.RefTypeK8sSecretKey, Name: "s", Key: "MANAGED"}, record.Variables["MANAGED"].Ref)
	})

	t.Run("unmanaged secret does not take an id", func(t *testing.T) {
		_, diags := completeEnvPlatformData(envPlatformDataInput{
			Values:    values(platformdata.PlatformAwsEcs),
			SecretIds: map[string]string{"MANAGED": "arn:m", "UNMANAGED": "arn:u"},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Invalid secret id key: UNMANAGED", diags[0].Summary)
	})

	t.Run("record passes through unchanged apart from refs", func(t *testing.T) {
		record, diags := completeEnvPlatformData(envPlatformDataInput{
			Values:    values(platformdata.PlatformAwsEcs),
			SecretIds: map[string]string{"MANAGED": "arn:m"},
		})
		require.Empty(t, diags)
		assert.Equal(t, platformdata.EnvV1Variable{Value: "x", Source: platformdata.SourceStandard}, record.Variables["PLAIN"])
		assert.Equal(t, platformdata.EnvV1Variable{Sensitive: true, Ref: &platformdata.EnvV1Ref{Type: platformdata.RefTypeSecret, Id: "arn:m"}, Source: platformdata.SourceUser}, record.Variables["MANAGED"])
		assert.Equal(t, "existing", record.Variables["UNMANAGED"].Ref.Id)
	})
}
