package provider

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/http"
)

func mockNsServerWithSubdomains() http.Handler {
	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/stacks/{stackName}/subdomains/{subdomainName}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subdomain := types.Subdomain{
				DnsName: "api",
				Block: types.Block{
					IdModel:   types.IdModel{Id: 123},
					OrgName:   "org0",
					StackId:   100,
					StackName: "demo",
				},
			}
			raw, _ := json.Marshal(subdomain)
			w.Write(raw)
		})
	return router
}
