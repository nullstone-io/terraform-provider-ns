package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataWorkspace(t *testing.T) {
	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.ns_workspace.this", `tags.%`, "3"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", `tags.Stack`, "stack0"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", `tags.Env`, "env0"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", `tags.Block`, "block0"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", "hyphenated_name", "stack0-env0-block0"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", "slashed_name", "stack0/env0/block0"),
	)

	t.Run("sets up attributes properly hard-coded", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_workspace" "this" {
  stack = "stack0"
  env   = "env0"
  block = "block0"
}
`)
		getTfeConfig, _ := mockTfe(nil)

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getTfeConfig),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks,
				},
			},
		})
	})

	t.Run("sets up attributes properly from env vars", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_workspace" "this" {}
`)
		getTfeConfig, _ := mockTfe(nil)

		os.Setenv("NULLSTONE_STACK", "stack0")
		os.Setenv("NULLSTONE_ENV", "env0")
		os.Setenv("NULLSTONE_BLOCK", "block0")

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getTfeConfig),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks,
				},
			},
		})
	})
}
