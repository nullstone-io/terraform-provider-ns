package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"testing"
)

func TestResourceAutogenSubdomain(t *testing.T) {
	subdomains := map[string]map[string]*types.AutogenSubdomain{
		"org0": {},
	}
	delegations := map[string]map[string]*types.AutogenSubdomainDelegation{
		"org0": {},
	}

	t.Run("creates successfully", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
resource "ns_autogen_subdomain" "subdomain" {}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(subdomains, delegations))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ns_autogen_subdomain.subdomain", `name`, "xyz123"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain.subdomain", `domain_name`, "nullstone.app"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain.subdomain", `fqdn`, "xyz123.nullstone.app."),
		)
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig),
			Steps: []resource.TestStep{
				{
					Config: tfconfig,
					Check:  checks,
				},
			},
		})
	})
}
