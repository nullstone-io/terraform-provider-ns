package provider

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nullstone-io/module/config"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"os"
	"regexp"
	"testing"
)

func TestDataAppConnection(t *testing.T) {
	os.Setenv("NULLSTONE_STACK_ID", "100")
	os.Setenv("NULLSTONE_STACK_NAME", "stack0")
	os.Setenv("NULLSTONE_BLOCK_ID", "101")
	os.Setenv("NULLSTONE_BLOCK_NAME", "faceless")
	os.Setenv("NULLSTONE_ENV_ID", "102")
	os.Setenv("NULLSTONE_ENV_NAME", "env0")
	uid1 := uuid.New()
	uid2 := uuid.New()
	uid3 := uuid.New()
	uid5 := uuid.New()
	// app
	facelessEnv0 := types.Workspace{
		UidCreatedModel: types.UidCreatedModel{Uid: uid1},
		OrgName:         "org0",
		StackId:         100,
		StackName:       "stack0",
		BlockId:         101,
		BlockName:       "faceless",
		EnvId:           102,
		EnvName:         "env0",
	}
	// cluster
	lycanEnv0 := types.Workspace{
		UidCreatedModel: types.UidCreatedModel{Uid: uid2},
		OrgName:         "org0",
		StackId:         100,
		StackName:       "stack0",
		BlockId:         103,
		BlockName:       "lycan",
		EnvId:           102,
		EnvName:         "env0",
	}
	// network
	rikiEnv0 := types.Workspace{
		UidCreatedModel: types.UidCreatedModel{Uid: uid3},
		OrgName:         "org0",
		StackId:         100,
		StackName:       "stack0",
		BlockId:         105,
		BlockName:       "rikimaru",
		EnvId:           102,
		EnvName:         "env0",
	}
	// enigma (app) => faceless (app) => lycan (cluster) => riki (network)
	enigmaEnv0 := types.Workspace{
		UidCreatedModel: types.UidCreatedModel{Uid: uid5},
		OrgName:         "org0",
		StackId:         100,
		StackName:       "stack0",
		BlockId:         107,
		BlockName:       "enigma",
		EnvId:           102,
		EnvName:         "env0",
	}
	workspaces := []types.Workspace{facelessEnv0, lycanEnv0, rikiEnv0}
	runConfigs := map[string]types.RunConfig{
		uid1.String(): {
			WorkspaceUid: uid1,
			Connections: map[string]types.Connection{
				"cluster": {
					Connection: config.Connection{
						Type:     "cluster/aws-fargate",
						Optional: false,
					},
					Target: "lycan",
					Reference: &types.ConnectionTarget{
						StackId: lycanEnv0.StackId,
						BlockId: lycanEnv0.BlockId,
					},
					Unused: false,
				},
			},
		},
		uid2.String(): {
			WorkspaceUid: uid2,
			Connections: map[string]types.Connection{
				"network": {
					Connection: config.Connection{
						Type:     "network/aws",
						Optional: false,
					},
					Target: "rikimaru",
					Reference: &types.ConnectionTarget{
						StackId: rikiEnv0.StackId,
						BlockId: rikiEnv0.BlockId,
					},
					Unused: false,
				},
			},
		},
	}

	t.Run("fails when required and connection is not configured", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_app_connection" "network" {
  name = "network"
  type = "network/aws"
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(workspaces, runConfigs))
		defer closeNsFn()
		getTfeConfig, _ := mockTfe(nil)

		checks := resource.ComposeTestCheckFunc()
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig, nil),
			Steps: []resource.TestStep{
				{
					Config:      tfconfig,
					Check:       checks,
					ExpectError: regexp.MustCompile(`The connection "network" is missing.`),
				},
			},
		})
	})

	t.Run("finds app connection directly", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_app_connection" "cluster" {
  name = "cluster"
  type = "cluster/aws-fargate"
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `workspace_id`, "100/103/102"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test2`, "2"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key2`, "value2"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key3`, "value3"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(workspaces, runConfigs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(mockStateServerWith(enigmaEnv0, lycanEnv0, rikiEnv0))
		defer closeTfeFn()

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

	t.Run("finds transitive app connection", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_app_connection" "cluster" {
  name = "cluster"
  type = "cluster/aws-fargate"
}

data "ns_app_connection" "network" {
  name = "network"
  type = "network/aws"
  via  = data.ns_app_connection.cluster.name
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `workspace_id`, "100/103/102"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test2`, "2"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key2`, "value2"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key3`, "value3"),
			resource.TestCheckResourceAttr("data.ns_app_connection.network", `workspace_id`, "100/105/102"),
			resource.TestCheckResourceAttr("data.ns_app_connection.network", `outputs.placeholder`, "value"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(workspaces, runConfigs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(mockStateServerWith(enigmaEnv0, lycanEnv0, rikiEnv0))
		defer closeTfeFn()

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

	t.Run("finds transitive app connection with optional via", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}

data "ns_app_connection" "cluster" {
  name     = "cluster"
  type     = "cluster/aws-fargate"
  optional = true
}

data "ns_app_connection" "network" {
  name = "network"
  type = "network/aws"
  via  = data.ns_app_connection.cluster.name
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `workspace_id`, "100/103/102"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test2`, "2"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key2`, "value2"),
			resource.TestCheckResourceAttr("data.ns_app_connection.cluster", `outputs.test3.key3`, "value3"),
			resource.TestCheckResourceAttr("data.ns_app_connection.network", `workspace_id`, "100/105/102"),
			resource.TestCheckResourceAttr("data.ns_app_connection.network", `outputs.placeholder`, "value"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(workspaces, runConfigs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(mockStateServerWith(enigmaEnv0, lycanEnv0, rikiEnv0))
		defer closeTfeFn()

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
