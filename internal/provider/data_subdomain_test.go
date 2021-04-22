package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"net/http"
	"regexp"
	"testing"
)

func TestDataSubdomain(t *testing.T) {
	t.Run("fails to find non-existent subdomain", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_subdomain" "subdomain" {
  stack = "demo"
  block = "api-subdomain"
}
`)

		checks := resource.ComposeTestCheckFunc()

		getNsConfig, closeNsFn := mockNs(http.NotFoundHandler())
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig),
			Steps: []resource.TestStep{
				{
					Config: tfconfig,
					Check:  checks,
					ExpectError: regexp.MustCompile(`The subdomain in the stack "demo" and block "api-subdomain" does not exist in nullstone.`),
				},
			},
		})
	})

	t.Run("sets up attributes properly", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_subdomain" "subdomain" {
  stack = "demo"
  block = "api-subdomain"
}
`)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_subdomain.subdomain", `stack`, "demo"),
			resource.TestCheckResourceAttr("data.ns_subdomain.subdomain", `block`, "api-subdomain"),
			resource.TestCheckResourceAttr("data.ns_subdomain.subdomain", `dns_name`, "api"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithSubdomains())
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