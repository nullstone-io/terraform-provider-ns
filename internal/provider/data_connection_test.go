package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataConnection(t *testing.T) {
	t.Run("fails when required and connection is not configured", func(t *testing.T) {
		config := fmt.Sprintf(`
data "ns_connection" "network" {
  name = "network"
  type = "aws/network"
}
`)
		checks := resource.ComposeTestCheckFunc()

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					Check:       checks,
					ExpectError: regexp.MustCompile(`The connection "network" is missing.`),
				},
			},
		})
	})

	t.Run("sets empty attributes when optional and connection is not configured", func(t *testing.T) {
		config := fmt.Sprintf(`
data "ns_connection" "service" {
  name     = "service"
  type     = "fargate/service"
  optional = true
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.service", `workspace`, ""),
		)

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks,
				},
			},
		})

	})

	t.Run("sets up attributes properly", func(t *testing.T) {
		config := fmt.Sprintf(`
data "ns_connection" "service" {
  name = "service"
  type = "fargate/service"
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.service", `workspace`, "lycan"),
		)

		os.Setenv("NULLSTONE_CONNECTION_service", "lycan")

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks,
				},
			},
		})
	})
}
