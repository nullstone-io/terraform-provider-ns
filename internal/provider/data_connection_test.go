package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataConnection(t *testing.T) {
	t.Run("fails when required and connection is not configured", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_connection" "network" {
  name = "network"
  type = "aws/network"
}
`)
		checks := resource.ComposeTestCheckFunc()
		getTfeConfig, _ := mockTfe(nil)

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getTfeConfig),
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
provider "ns" {
  organization = "org0"
}
data "ns_connection" "service" {
  name     = "service"
  type     = "fargate/service"
  optional = true
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.service", `workspace`, ""),
			resource.TestCheckResourceAttr("data.ns_connection.service", `outputs.%`, "0"),
		)
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

	t.Run("sets up attributes properly", func(t *testing.T) {
		config := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}
data "ns_connection" "service" {
  name = "service"
  type = "fargate/service"
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_connection.service", `workspace`, "stack0-env0-lycan"),
			resource.TestCheckResourceAttr("data.ns_connection.service", `outputs.test1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.service", `outputs.test2`, "2"),
			resource.TestCheckResourceAttr("data.ns_connection.service", `outputs.test3.key1`, "value1"),
			resource.TestCheckResourceAttr("data.ns_connection.service", `outputs.test3.key2`, "value2"),
			resource.TestCheckResourceAttr("data.ns_connection.service", `outputs.test3.key3`, "value3"),
		)

		os.Setenv("NULLSTONE_CONNECTION_service", "stack0-env0-lycan")

		getTfeConfig, closeFn := mockTfe(mockServerWithLycan())
		defer closeFn()

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

func mockServerWithLycan() http.Handler {
	workspaces := map[string]json.RawMessage{
		"stack0-env0-lycan": json.RawMessage(`{
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
	}
	currentStateVersions := map[string]json.RawMessage{
		"cb30d6ab-1a9e-4c7c-aaf2-9dc9f33eeabc": json.RawMessage(`{
  "data": {
    "id": "cb30d6ab-1a9e-4c7c-aaf2-9dc9f33eeabc",
    "type": "state-versions",
    "attributes": {
      "name": "stack0-env0-lycan",
      "serial": 1,
      "lineage": "64aef234-2ff9-9d8e-25ae-22fb30b62860",
      "hosted-state-download-url": "/state/terraform/v2/state-versions/53516a9e-ffd7-4834-8234-63fd070d064f/download"
    },
    "relationships": {}
  }
}`),
	}
	stateFiles := map[string]json.RawMessage{
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
	}
	return mockTfeStatePull(workspaces, currentStateVersions, stateFiles)
}

func mockTfeStatePull(workspaces map[string]json.RawMessage, currentStateVersions map[string]json.RawMessage, stateFiles map[string]json.RawMessage) http.Handler {
	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/state/terraform/v2/organizations/{orgName}/workspaces/{workspaceName}").
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
		Path("/state/terraform/v2/workspaces/{workspaceId}/current-state-version").
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
		Path("/state/terraform/v2/state-versions/{stateVersionId}/download").
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
