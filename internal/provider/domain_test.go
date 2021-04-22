package provider

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/http"
)

func mockNsServerWithDomains() http.Handler {
	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/stacks/{stackName}/domains/{domainName}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			domain := types.Domain{
				DnsName: "nullstone.io",
				OrgName: "org0",
				StackName: "global",
			}
			raw, _ := json.Marshal(domain)
			w.Write(raw)
		})
	return router
}
