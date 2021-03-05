package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"regexp"
	"testing"
)

func TestResourceSubdomainDelegation(t *testing.T) {
	subdomains := map[string]map[string]*ns.AutogenSubdomain{
		"org0": {
			"api": {
				Id:         1,
				Name:       "api",
				DomainName: "nullstone.app",
			},
			"docs": {
				Id:         2,
				Name:       "docs",
				DomainName: "nullstone.app",
			},
		},
	}
	delegations := map[string]map[string]*ns.AutogenSubdomainDelegation{
		"org0": {
			"docs": {
				Nameservers: []string{"2.2.2.2", "3.3.3.3", "4.4.4.4"},
			},
		},
	}

	t.Run("fails to update non-existent delegation", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
resource "ns_autogen_subdomain_delegation" "to_fake" {
  subdomain   = "missing"
  nameservers = ["1.1.1.1","2.2.2.2","3.3.3.3"]
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
					ExpectError: regexp.MustCompile(`The autogen_subdomain_delegation "missing" is missing.`),
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
  subdomain   = "api"
  nameservers = ["1.1.1.1","2.2.2.2","3.3.3.3"]
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(subdomains, delegations))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `subdomain`, "api"),
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
  subdomain   = "docs"
  nameservers = ["5.5.5.5", "6.6.6.6", "7.7.7.7"]
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAutogenSubdomains(subdomains, delegations))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("ns_autogen_subdomain_delegation.to_fake", `subdomain`, "docs"),
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
