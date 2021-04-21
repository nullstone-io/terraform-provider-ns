package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"regexp"
	"testing"
)

func TestDataAutogenSubdomain(t *testing.T) {
	subdomains := map[string]map[string]*types.AutogenSubdomain{
		"org0": {
			"api": {
				IdModel:    types.IdModel{Id: 1},
				Name:       "api",
				DomainName: "nullstone.app",
			},
		},
	}
	delegations := map[string]map[string]*types.AutogenSubdomainDelegation{}

	t.Run("fails to find non-existent autogen_subdomain", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_autogen_subdomain" "subdomain" {
  name = "docs"
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(subdomains, delegations))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc()
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig),
			Steps: []resource.TestStep{
				{
					Config:      tfconfig,
					Check:       checks,
					ExpectError: regexp.MustCompile(`The autogen_subdomain "docs" is missing.`),
				},
			},
		})
	})

	t.Run("sets up attributes properly", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_autogen_subdomain" "subdomain" {
  name = "api"
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_autogen_subdomain.subdomain", `name`, "api"),
			resource.TestCheckResourceAttr("data.ns_autogen_subdomain.subdomain", `domain_name`, "nullstone.app"),
			resource.TestCheckResourceAttr("data.ns_autogen_subdomain.subdomain", `fqdn`, "api.nullstone.app."),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(subdomains, delegations))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

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
