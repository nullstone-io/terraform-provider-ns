package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func mockNsServerWithAgent(agent types.NullstoneAgent) http.Handler {
	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/nullstone_agent").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(agent)
		})
	return router
}

func TestDataAgent(t *testing.T) {
	tfconfig := `
provider "ns" {
  organization = "org0"
}
data "ns_agent" "this" {
}
`

	t.Run("sets up attributes properly", func(t *testing.T) {
		agent := types.NullstoneAgent{
			Aws: types.NullstoneAgentInfoAws{
				AccountId: "123456789012",
				UserName:  "execution-assumer",
				UserArn:   "arn:aws:iam::123456789012:user/execution-assumer",
			},
			Gcp: types.NullstoneAgentInfoGcp{
				ProjectId:           "assumer",
				ServiceAccountEmail: "nullstone-assumer@assumer.iam.gserviceaccount.com",
			},
			IpAddresses: []string{"203.0.113.10", "203.0.113.11"},
		}

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_agent.this", `aws_account_id`, "123456789012"),
			resource.TestCheckResourceAttr("data.ns_agent.this", `aws_user_name`, "execution-assumer"),
			resource.TestCheckResourceAttr("data.ns_agent.this", `aws_user_arn`, "arn:aws:iam::123456789012:user/execution-assumer"),
			resource.TestCheckResourceAttr("data.ns_agent.this", `gcp_project_id`, "assumer"),
			resource.TestCheckResourceAttr("data.ns_agent.this", `gcp_service_account_email`, "nullstone-assumer@assumer.iam.gserviceaccount.com"),
			resource.TestCheckResourceAttr("data.ns_agent.this", `ip_addresses.#`, "2"),
			resource.TestCheckResourceAttr("data.ns_agent.this", `ip_addresses.0`, "203.0.113.10"),
			resource.TestCheckResourceAttr("data.ns_agent.this", `ip_addresses.1`, "203.0.113.11"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAgent(agent))
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

	t.Run("ip_addresses is an empty list when the API omits it", func(t *testing.T) {
		agent := types.NullstoneAgent{
			Aws: types.NullstoneAgentInfoAws{AccountId: "123456789012"},
		}

		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_agent.this", `aws_account_id`, "123456789012"),
			resource.TestCheckResourceAttr("data.ns_agent.this", `ip_addresses.#`, "0"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsServerWithAgent(agent))
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
