package provider

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/http"
)

func mockNsServerWithSubdomains() http.Handler {
	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/subdomains/{subdomainId}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subdomain := types.Subdomain{
				Block: types.Block{
					IdModel: types.IdModel{Id: 123},
					OrgName: "org0",
					StackId: 100,
					DnsName: "api",
				},
			}
			raw, _ := json.Marshal(subdomain)
			w.Write(raw)
		})
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/stacks/{stackId}/subdomains/{subdomainId}/envs/{envId}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subdomainWorkspace := types.SubdomainWorkspace{
				WorkspaceUid:  uuid.UUID{},
				DnsName:       "api",
				SubdomainName: "api.dev",
				DomainName:    "acme.com",
				Fqdn:          "api.dev.acme.com.",
			}
			raw, _ := json.Marshal(subdomainWorkspace)
			w.Write(raw)
		})
	return router
}
