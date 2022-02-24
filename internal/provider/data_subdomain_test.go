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
  stack_id = 100
  block_id = 126
}
`)

		checks := resource.ComposeTestCheckFunc()

		getNsConfig, closeNsFn := mockNs(http.NotFoundHandler())
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig, nil),
			Steps: []resource.TestStep{
				{
					Config:      tfconfig,
					Check:       checks,
					ExpectError: regexp.MustCompile(`The subdomain in the stack 100 and block 126 does not exist in nullstone.`),
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
  stack_id = 100
  block_id = 123
}
`)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_subdomain.subdomain", `stack_id`, "100"),
			resource.TestCheckResourceAttr("data.ns_subdomain.subdomain", `block_id`, "123"),
			resource.TestCheckResourceAttr("data.ns_subdomain.subdomain", `dns_name`, "api"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithSubdomains())
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

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
