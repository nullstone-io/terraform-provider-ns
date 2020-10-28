package ns

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccNsWorkspaceDataSource(t *testing.T) {
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
data "ns_workspace" "this" {
  stack = "stack0"
  env   = "env0"
  block = "block0"
}
`)
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
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
data "ns_workspace" "this" {}
`)

		os.Setenv("NULLSTONE_STACK", "stack0")
		os.Setenv("NULLSTONE_ENV", "env0")
		os.Setenv("NULLSTONE_BLOCK", "block0")

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks,
				},
			},
		})
	})
}
