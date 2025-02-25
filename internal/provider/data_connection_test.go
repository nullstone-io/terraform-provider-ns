package provider

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nullstone-io/module/config"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/workspaces"
	"net/http"
	"os"
	"regexp"
	"testing"
)

func TestDataConnection(t *testing.T) {
	os.Setenv("NULLSTONE_STACK_ID", "100")
	os.Setenv("NULLSTONE_STACK_NAME", "stack0")
	os.Setenv("NULLSTONE_BLOCK_ID", "101")
	os.Setenv("NULLSTONE_BLOCK_NAME", "faceless")
	os.Setenv("NULLSTONE_ENV_ID", "102")
	os.Setenv("NULLSTONE_ENV_NAME", "env0")
	uid1 := uuid.New()
	uid2 := uuid.New()
	uid3 := uuid.New()
	uid4 := uuid.New()
	uid5 := uuid.New()
	// faceless (app) => lycan (cluster) => riki (network)
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
	// techies (app) => (no connection) - set up to test plan config
	techiesEnv0 := types.Workspace{
		UidCreatedModel: types.UidCreatedModel{Uid: uid4},
		OrgName:         "org0",
		StackId:         100,
		StackName:       "stack0",
		BlockId:         106,
		BlockName:       "techies",
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
	allWorkspaces := []types.Workspace{facelessEnv0, lycanEnv0, rikiEnv0, techiesEnv0, enigmaEnv0}
	runConfigs := map[string]types.RunConfig{
		uid1.String(): {
			WorkspaceUid: uid1,
			WorkspaceConfig: types.WorkspaceConfig{
				Connections: map[string]types.Connection{
					"cluster": {
						Connection: config.Connection{
							Type:     "cluster/aws-fargate",
							Contract: "cluster/aws/ecs",
							Optional: false,
						},
						DesiredTarget: &types.ConnectionTarget{BlockName: "lycan"},
						EffectiveTarget: &types.ConnectionTarget{
							StackId: lycanEnv0.StackId,
							BlockId: lycanEnv0.BlockId,
						},
						Unused: false,
					},
				},
			},
		},
		uid2.String(): {
			WorkspaceUid: uid2,
			WorkspaceConfig: types.WorkspaceConfig{
				Connections: map[string]types.Connection{
					"network": {
						Connection: config.Connection{
							Type:     "network/aws",
							Contract: "network/aws/vpc",
							Optional: false,
						},
						DesiredTarget: &types.ConnectionTarget{BlockName: "rikimaru"},
						EffectiveTarget: &types.ConnectionTarget{
							StackId: rikiEnv0.StackId,
							BlockId: rikiEnv0.BlockId,
						},
						Unused: false,
					},
				},
			},
		},
		uid4.String(): {
			WorkspaceUid: uid4,
			WorkspaceConfig: types.WorkspaceConfig{
				Connections: map[string]types.Connection{
					// Intentionally blank, this simulates a connection not configured yet on Nullstone servers
				},
			},
		},
		uid5.String(): {
			WorkspaceUid: uid5,
			WorkspaceConfig: types.WorkspaceConfig{
				Connections: map[string]types.Connection{
					"app": {
						Connection: config.Connection{
							Contract: "app/aws/ecs",
							Optional: false,
						},
						DesiredTarget: &types.ConnectionTarget{BlockName: "faceless"},
						EffectiveTarget: &types.ConnectionTarget{
							StackId: facelessEnv0.StackId,
							BlockId: facelessEnv0.BlockId,
						},
						Unused: false,
					},
				},
			},
		},
	}

	t.Run("fails when required and connection is not configured", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws/vpc"
}
`)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(allWorkspaces, runConfigs))
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

	t.Run("sets empty attributes when optional and connection is not configured", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_connection" "postgres" {
  name     = "postgres"
  contract = "datastore/aws/postgres:rds"
  optional = true
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.postgres", `workspace_id`, ""),
			resource.TestCheckResourceAttr("data.ns_connection.postgres", `outputs.%`, "0"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(allWorkspaces, runConfigs))
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

	t.Run("sets up attributes properly", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_connection" "cluster" {
  name     = "cluster"
  contract = "cluster/aws/ecs"
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `workspace_id`, "100/103/102"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test2`, "2"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key2`, "value2"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key3`, "value3"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(allWorkspaces, runConfigs))
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

	t.Run("sets up attributes with via properly", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_connection" "cluster" {
  name     = "cluster"
  contract = "cluster/aws/ecs"
}
data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws/vpc"
  via      = data.ns_connection.cluster.name
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `workspace_id`, "100/103/102"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test2`, "2"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key2`, "value2"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key3`, "value3"),
			resource.TestCheckResourceAttr("data.ns_connection.network", `workspace_id`, "100/105/102"),
			resource.TestCheckResourceAttr("data.ns_connection.network", `outputs.placeholder`, "value"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(allWorkspaces, runConfigs))
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

	t.Run("sets up attributes with transitive via properly", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_connection" "app" {
  name     = "app"
  contract = "app/aws/ecs"
}
data "ns_connection" "cluster" {
  name     = "cluster"
  contract = "cluster/aws/ecs"
  via      = data.ns_connection.app.name
}
data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws/vpc"
  via      = "${data.ns_connection.app.name}/${data.ns_connection.cluster.name}"
}
`)

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.app", `workspace_id`, "100/101/102"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `workspace_id`, "100/103/102"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test2`, "2"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key2`, "value2"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key3`, "value3"),
			resource.TestCheckResourceAttr("data.ns_connection.network", `outputs.placeholder`, "value"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(allWorkspaces, runConfigs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(mockStateServerWith(enigmaEnv0, lycanEnv0, rikiEnv0))
		defer closeTfeFn()

		// We alter the plan config to set the workspace context to enigma
		//   so that the app can satisfy app connection with faceless
		alterPlanConfig := func(config *PlanConfig) {
			config.BlockId = enigmaEnv0.BlockId
			config.BlockName = enigmaEnv0.BlockName
		}

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig, alterPlanConfig),
			Steps: []resource.TestStep{
				{
					Config: tfconfig,
					Check:  checks,
				},
			},
		})
	})

	t.Run("overrides connection through local setup in plan config", func(t *testing.T) {
		// This simulates connections injected via .nullstone/active-workspace.yml
		// Later in the test, we configure the plan config with: techies (app) => lycan (cluster)
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_connection" "cluster" {
  name     = "cluster"
  contract = "cluster/aws/ecs"
}
data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws/vpc"
  via      = data.ns_connection.cluster.name
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `workspace_id`, "100/103/102"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test2`, "2"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key2`, "value2"),
			resource.TestCheckResourceAttr("data.ns_connection.cluster", `outputs.test3.key3`, "value3"),
			resource.TestCheckResourceAttr("data.ns_connection.network", `workspace_id`, "100/105/102"),
			resource.TestCheckResourceAttr("data.ns_connection.network", `outputs.placeholder`, "value"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWith(allWorkspaces, runConfigs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(mockStateServerWith(enigmaEnv0, lycanEnv0, rikiEnv0))
		defer closeTfeFn()
		alterPlanConfig := func(config *PlanConfig) {
			// Set up cluster connection for techies
			config.OrgName = techiesEnv0.OrgName
			config.StackId = techiesEnv0.StackId
			config.BlockId = techiesEnv0.BlockId
			config.EnvId = techiesEnv0.EnvId
			config.WorkspaceUid = techiesEnv0.Uid.String()
			config.Connections = workspaces.ManifestConnections{
				"cluster": { // point at lycan
					StackId: lycanEnv0.StackId,
					BlockId: lycanEnv0.BlockId,
				},
			}
		}

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig, alterPlanConfig),
			Steps: []resource.TestStep{
				{
					Config: tfconfig,
					Check:  checks,
				},
			},
		})

	})
}

func mockNsServerWith(workspaces []types.Workspace, runConfigs map[string]types.RunConfig) http.Handler {
	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/stacks/{stackId}/blocks/{blockId}/envs/{envId}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, stackId := vars["orgName"], vars["stackId"]
			blockId, envId := vars["blockId"], vars["envId"]
			for _, workspace := range workspaces {
				if workspace.OrgName == orgName &&
					fmt.Sprintf("%d", workspace.StackId) == stackId &&
					fmt.Sprintf("%d", workspace.BlockId) == blockId &&
					fmt.Sprintf("%d", workspace.EnvId) == envId {
					raw, _ := json.Marshal(workspace)
					w.Write(raw)
					return
				}
			}
			http.NotFound(w, r)
		})
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/stacks/{stackId}/workspaces/{workspaceUid}/run-configs/latest").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, stackId, workspaceUidStr := vars["orgName"], vars["stackId"], vars["workspaceUid"]
			for _, workspace := range workspaces {
				if workspace.OrgName == orgName && fmt.Sprintf("%d", workspace.StackId) == stackId {
					if workspace.Uid.String() != workspaceUidStr {
						continue
					} else if rc, ok := runConfigs[workspaceUidStr]; !ok {
						http.NotFound(w, r)
					} else {
						raw, _ := json.Marshal(rc)
						w.Write(raw)
					}
					return
				}
			}

			http.NotFound(w, r)
		})
	return router
}

func mockStateServerWith(enigmaEnv0 types.Workspace, lycanEnv0 types.Workspace, rikiEnv0 types.Workspace) http.Handler {
	workspaces := map[string]json.RawMessage{
		enigmaEnv0.Uid.String(): json.RawMessage(`{
  "data": {
    "id": "cb30d6ab-2345-4c7c-aaf2-9dc9f33eeabc",
    "type": "workspaces",
    "attributes": {
      "name": "stack0-env0-enigma"
    },
    "relationships": {
      "organization": {
        "data": {
          "id": "org0",
          "type": "organizations"
        }
      }
    }
  }
}`),
		lycanEnv0.Uid.String(): json.RawMessage(`{
  "data": {
    "id": "cb30d6ab-1a9e-4c7c-aaf2-9dc9f33eeabc",
    "type": "workspaces",
    "attributes": {
      "name": "stack0-env0-lycan"
    },
    "relationships": {
      "organization": {
        "data": {
          "id": "org0",
          "type": "organizations"
        }
      }
    }
  }
}`),
		rikiEnv0.Uid.String(): json.RawMessage(`{
  "data": {
    "id": "ce69c4d8-5c90-41ab-a0ba-3ef770efbdb1",
    "type": "workspaces",
    "attributes": {
      "name": "stack0-env0-rikimaru"
    },
    "relationships": {
      "organization": {
        "data": {
          "id": "org0",
          "type": "organizations"
        }
      }
    }
  }
}`),
	}
	currentStateVersions := map[string]json.RawMessage{
		"b6876f4b-7615-4197-abd8-93835ed175bf": json.RawMessage(`{
  "data": {
    "id": "b2d5fc01-e51c-462b-bff5-9130d31ebc71",
    "type": "state-versions",
    "attributes": {
      "name": "stack0-env0-enigma",
      "serial": 1,
      "lineage": "0b4b9d07-cad7-4acd-8ce9-40639b46f3b0",
      "hosted-state-download-url": "/terraform/v2/state-versions/b2d5fc01-e51c-462b-bff5-9130d31ebc71/download"
    },
    "relationships": {}
  }
}`),
		"cb30d6ab-1a9e-4c7c-aaf2-9dc9f33eeabc": json.RawMessage(`{
  "data": {
    "id": "53516a9e-ffd7-4834-8234-63fd070d064f",
    "type": "state-versions",
    "attributes": {
      "name": "stack0-env0-lycan",
      "serial": 1,
      "lineage": "64aef234-2ff9-9d8e-25ae-22fb30b62860",
      "hosted-state-download-url": "/terraform/v2/state-versions/53516a9e-ffd7-4834-8234-63fd070d064f/download"
    },
    "relationships": {}
  }
}`),
		"ce69c4d8-5c90-41ab-a0ba-3ef770efbdb1": json.RawMessage(`{
  "data": {
    "id": "007eb553-710a-49fb-8ded-2c342702d6b3",
    "type": "state-versions",
    "attributes": {
      "name": "stack0-env0-rikimaru",
      "serial": 1,
      "lineage": "bc6743fd-f886-4050-a2a1-d5fa66c0e22a",
      "hosted-state-download-url": "/terraform/v2/state-versions/007eb553-710a-49fb-8ded-2c342702d6b3/download"
    },
    "relationships": {}
  }
}`),
	}
	stateFiles := map[string]json.RawMessage{
		"b2d5fc01-e51c-462b-bff5-9130d31ebc71": json.RawMessage(`{
  "version": 4,
  "terraform_version": "0.13.5",
  "serial": 1,
  "lineage": "0b4b9d07-cad7-4acd-8ce9-40639b46f3b0",
  "outputs": {}
  },
  "resources": []
}`),
		"53516a9e-ffd7-4834-8234-63fd070d064f": json.RawMessage(`{
  "version": 4,
  "terraform_version": "0.13.5",
  "serial": 1,
  "lineage": "64aef234-2ff9-9d8e-25ae-22fb30b62860",
  "outputs": {
    "test1": {
      "value": "value1",
      "type": "string"
    },
    "test2": {
      "value": 2,
      "type": "number"
    },
    "test3": {
      "value": {
        "key1": "value1",
        "key2": "value2",
        "key3": "value3"
      },
      "type": [
        "object",
        {
          "key1": "string",
          "key2": "string",
          "key3": "string"
        }
      ]
    }
  },
  "resources": []
}`),
		"007eb553-710a-49fb-8ded-2c342702d6b3": json.RawMessage(`{
  "version": 4,
  "terraform_version": "0.13.5",
  "serial": 1,
  "lineage": "bc6743fd-f886-4050-a2a1-d5fa66c0e22a",
  "outputs": {
    "placeholder": {
      "value": "value",
      "type": "string"
    }
  },
  "resources": []
}`),
	}
	return mockTfeStatePull(workspaces, currentStateVersions, stateFiles)
}

func mockTfeStatePull(workspaces map[string]json.RawMessage, currentStateVersions map[string]json.RawMessage, stateFiles map[string]json.RawMessage) http.Handler {
	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/terraform/v2/organizations/{orgName}/workspaces/{workspaceName}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			_, workspaceName := vars["orgName"], vars["workspaceName"]
			if msg, ok := workspaces[workspaceName]; ok {
				w.Write(msg)
			} else {
				http.NotFound(w, r)
			}
		})
	router.
		Methods(http.MethodGet).
		Path("/terraform/v2/workspaces/{workspaceId}/current-state-version").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			workspaceId := mux.Vars(r)["workspaceId"]
			if msg, ok := currentStateVersions[workspaceId]; ok {
				w.Write(msg)
			} else {
				http.NotFound(w, r)
			}
		})
	router.
		Methods(http.MethodGet).
		Path("/terraform/v2/state-versions/{stateVersionId}/download").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stateVersionId := mux.Vars(r)["stateVersionId"]
			if msg, ok := stateFiles[stateVersionId]; ok {
				w.Write(msg)
			} else {
				http.NotFound(w, r)
			}
		})
	return router
}
