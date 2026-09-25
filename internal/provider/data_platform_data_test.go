package provider

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataPlatformData(t *testing.T) {
	getNsConfig, _ := mockNs(nil)
	getTfeConfig, _ := mockTfe(nil)
	factories := protoV5ProviderFactories(getNsConfig, getTfeConfig, nil)

	t.Run("valid env v1 payload persists into state", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_platform_data" "env" {
  kind    = "env"
  version = 1
  data = jsonencode({
    variables = {
      NULLSTONE_ENV = { value = "dev" }
      DATABASE_NAME = { template = "{{ NULLSTONE_ENV }}-db", value = "dev-db" }
    }
    secret_keys = ["DATABASE_PASSWORD"]
  })
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.ns_platform_data.env", "id", "env/1"),
						resource.TestCheckResourceAttr("data.ns_platform_data.env", "kind", "env"),
						resource.TestCheckResourceAttr("data.ns_platform_data.env", "version", "1"),
						resource.TestCheckResourceAttrSet("data.ns_platform_data.env", "data"),
						resource.TestMatchResourceAttr("data.ns_platform_data.env", "data", regexp.MustCompile(`"DATABASE_PASSWORD"`)),
					),
				},
			},
		})
	})

	t.Run("invalid env v1 payload errors at plan", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_platform_data" "env" {
  kind    = "env"
  version = 1
  data = jsonencode({
    variables = {
      DATABASE_PASSWORD = { value = "leaked" }
    }
    secret_keys = ["DATABASE_PASSWORD"]
  })
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`cannot be both a\s+variable and a secret`),
				},
			},
		})
	})

	t.Run("unknown kind is stored with a warning", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_platform_data" "metrics" {
  kind    = "metrics"
  version = 1
  data = jsonencode({
    endpoints = ["/metrics"]
  })
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.ns_platform_data.metrics", "id", "metrics/1"),
						resource.TestCheckResourceAttr("data.ns_platform_data.metrics", "kind", "metrics"),
						resource.TestCheckResourceAttr("data.ns_platform_data.metrics", "version", "1"),
					),
				},
			},
		})
	})

	t.Run("data that is not a JSON object errors", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_platform_data" "env" {
  kind    = "env"
  version = 1
  data    = jsonencode(["not", "an", "object"])
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`data must be a JSON-encoded object`),
				},
			},
		})
	})

	t.Run("non-integer version errors", func(t *testing.T) {
		config := `
provider "ns" {
  organization = "org0"
}
data "ns_platform_data" "env" {
  kind    = "env"
  version = 1.5
  data    = jsonencode({ variables = {} })
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: factories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`version must be a positive integer`),
				},
			},
		})
	})
}

func TestDataPlatformData_Validate(t *testing.T) {
	ds := &dataPlatformData{p: &provider{}}

	t.Run("skips validation when data is unknown", func(t *testing.T) {
		config := map[string]tftypes.Value{
			"kind":    tftypes.NewValue(tftypes.String, "env"),
			"version": tftypes.NewValue(tftypes.Number, 1),
			"data":    tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		}
		diags, err := ds.Validate(context.Background(), config)
		require.NoError(t, err)
		assert.Empty(t, diags)
	})

	t.Run("skips validation when kind is null", func(t *testing.T) {
		config := map[string]tftypes.Value{
			"kind":    tftypes.NewValue(tftypes.String, nil),
			"version": tftypes.NewValue(tftypes.Number, 1),
			"data":    tftypes.NewValue(tftypes.String, `{"variables":{}}`),
		}
		diags, err := ds.Validate(context.Background(), config)
		require.NoError(t, err)
		assert.Empty(t, diags)
	})

	t.Run("reports warning for unknown version", func(t *testing.T) {
		config := map[string]tftypes.Value{
			"kind":    tftypes.NewValue(tftypes.String, "env"),
			"version": tftypes.NewValue(tftypes.Number, 99),
			"data":    tftypes.NewValue(tftypes.String, `{"variables":{}}`),
		}
		diags, err := ds.Validate(context.Background(), config)
		require.NoError(t, err)
		require.Len(t, diags, 1)
		assert.Equal(t, "Unrecognized platform data", diags[0].Summary)
	})

	t.Run("reports error for invalid env payload", func(t *testing.T) {
		config := map[string]tftypes.Value{
			"kind":    tftypes.NewValue(tftypes.String, "env"),
			"version": tftypes.NewValue(tftypes.Number, 1),
			"data":    tftypes.NewValue(tftypes.String, `{"secret_keys":[]}`),
		}
		diags, err := ds.Validate(context.Background(), config)
		require.NoError(t, err)
		require.Len(t, diags, 1)
		assert.Equal(t, "Invalid platform data", diags[0].Summary)
		assert.Contains(t, diags[0].Detail, "variables is required")
	})
}
