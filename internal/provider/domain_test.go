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
		Path("/orgs/{orgName}/stacks/{stackId}/domains/{domainId}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			domain := types.Domain{
				Block: types.Block{
					IdModel: types.IdModel{Id: 117},
					OrgName: "org0",
					StackId: 100,
					DnsName: "nullstone.io",
				},
			}
			raw, _ := json.Marshal(domain)
			w.Write(raw)
		})
	return router
}
