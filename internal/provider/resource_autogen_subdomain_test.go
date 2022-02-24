package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"testing"
)

func TestResourceAutogenSubdomain(t *testing.T) {
	autogenSubdomains := map[string]map[string]map[string]*types.AutogenSubdomain{
		"org0": {},
	}

	t.Run("creates successfully", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
resource "ns_autogen_subdomain" "autogen_subdomain" {
  subdomain_id 	= 99
  env_id        = 15
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(autogenSubdomains))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ns_autogen_subdomain.autogen_subdomain", `dns_name`, "xyz123"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain.autogen_subdomain", `domain_name`, "nullstone.app"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain.autogen_subdomain", `fqdn`, "xyz123.nullstone.app."),
		)
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig, nil),
			Steps: []resource.TestStep{
				{
					Config: tfconfig,
					Check:  checks,
				},
			},
		})
	})
}
