package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"regexp"
	"testing"
)

func TestResourceSubdomainDelegation(t *testing.T) {
	autogenSubdomains := map[string]map[string]map[string]*types.AutogenSubdomain{
		"org0": {
			"1": {
				"prod": {
					IdModel:     types.IdModel{Id: 1},
					DnsName:     "api",
					DomainName:  "nullstone.app",
					Fqdn:        "api.nullstone.app.",
					Nameservers: []string{},
				},
			},
			"2": {
				"prod": {
					IdModel:     types.IdModel{Id: 2},
					DnsName:     "docs",
					DomainName:  "nullstone.app",
					Fqdn:        "docs.nullstone.app.",
					Nameservers: []string{"2.2.2.2", "3.3.3.3", "4.4.4.4"},
				},
			},
		},
	}

	t.Run("fails to update non-existent delegation", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
resource "ns_autogen_subdomain_delegation" "to_fake" {
  subdomain_id = 99
  env 		   = "prod"
  nameservers  = ["1.1.1.1","2.2.2.2","3.3.3.3"]
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(autogenSubdomains))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc()
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig),
			Steps: []resource.TestStep{
				{
					Config:      tfconfig,
					Check:       checks,
					ExpectError: regexp.MustCompile(`The autogen_subdomain_delegation for the subdomain 99 and env "prod" is missing.`),
				},
			},
		})
	})

	t.Run("correctly updates new delegation", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
resource "ns_autogen_subdomain_delegation" "to_fake" {
  subdomain_id	= 1
  env 			= "prod"
  nameservers 	= ["1.1.1.1","2.2.2.2","3.3.3.3"]
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(autogenSubdomains))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `nameservers.#`, "3"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `nameservers.0`, "1.1.1.1"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `nameservers.1`, "2.2.2.2"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `nameservers.2`, "3.3.3.3"),
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

	t.Run("correctly updates existing delegation", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
resource "ns_autogen_subdomain_delegation" "to_fake" {
  subdomain_id 	= 2
  env 			= "prod"
  nameservers 	= ["5.5.5.5", "6.6.6.6", "7.7.7.7"]
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(autogenSubdomains))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `nameservers.#`, "3"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `nameservers.0`, "5.5.5.5"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `nameservers.1`, "6.6.6.6"),
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `nameservers.2`, "7.7.7.7"),
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
