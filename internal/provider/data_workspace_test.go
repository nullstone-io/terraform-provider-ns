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
		resource.TestCheckResourceAttr("data.ns_workspace.this", "stack_id", "100"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", "stack_name", "stack0"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", "block_id", "101"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", "block_name", "block0"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", "block_ref", "yellow-giraffe"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", "env_id", "102"),
		resource.TestCheckResourceAttr("data.ns_workspace.this", "env_name", "env0"),
	)

	t.Run("sets up attributes properly hard-coded", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_workspace" "this" {
  stack_id   = "100"
  stack_name = "stack0"
  block_id   = "101"
  block_name = "block0"
  block_ref  = "yellow-giraffe"
  env_id     = "102"
  env_name   = "env0"
}
`)
		getNsConfig, _ := mockNs(nil)
		getTfeConfig, _ := mockTfe(nil)

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig, nil),
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
		getNsConfig, _ := mockNs(nil)
		getTfeConfig, _ := mockTfe(nil)

		os.Setenv("NULLSTONE_STACK_ID", "100")
		os.Setenv("NULLSTONE_STACK_NAME", "stack0")
		os.Setenv("NULLSTONE_BLOCK_ID", "101")
		os.Setenv("NULLSTONE_BLOCK_NAME", "block0")
		os.Setenv("NULLSTONE_BLOCK_REF", "yellow-giraffe")
		os.Setenv("NULLSTONE_ENV_ID", "102")
		os.Setenv("NULLSTONE_ENV_NAME", "env0")

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig, nil),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks,
				},
			},
		})
	})
}
