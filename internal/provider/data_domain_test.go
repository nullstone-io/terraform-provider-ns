package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"net/http"
	"regexp"
	"testing"
)

func TestDataDomain(t *testing.T) {
	t.Run("fails to find non-existent domain", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_domain" "domain" {
  stack = "global"
  block = "nullstone-io"
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
					Config:      tfconfig,
					Check:       checks,
					ExpectError: regexp.MustCompile(`The domain in the stack "global" and block "nullstone-io" does not exist in nullstone.`),
				},
			},
		})
	})

	t.Run("sets up attributes properly", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_domain" "domain" {
  stack = "global"
  block = "nullstone-io"
}
`)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_domain.domain", `stack`, "global"),
			resource.TestCheckResourceAttr("data.ns_domain.domain", `block`, "nullstone-io"),
			resource.TestCheckResourceAttr("data.ns_domain.domain", `dns_name`, "nullstone.io"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithDomains())
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
