package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataCapability(t *testing.T) {
	t.Run("reads capability from plan config", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization    = "org0"
  capability_id   = "123"
  capability_name = "my-capability"
}
data "ns_capability" "this" {}
`)
		getNsConfig, _ := mockNs(nil)
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_capability.this", "id", "123"),
			resource.TestCheckResourceAttr("data.ns_capability.this", "name", "my-capability"),
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
