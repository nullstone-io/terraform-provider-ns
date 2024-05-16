package provider

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/http"
)

func mockNsServerWithAutogenSubdomains(autogenSubdomains map[string]map[string]map[string]*types.AutogenSubdomain) http.Handler {
	indexFn := func(orgName, subdomainId, envId string) string {
		return fmt.Sprintf("%s/%s/%s", orgName, subdomainId, envId)
	}

	indexed := map[string]*types.AutogenSubdomain{}
	for org, orgScoped := range autogenSubdomains {
		for subdomainId, subScoped := range orgScoped {
			for envId, as := range subScoped {
				indexed[indexFn(org, subdomainId, envId)] = as
			}
		}
	}

	findAutogenSubdomain := func(orgName string, subdomainId string, envId string) *types.AutogenSubdomain {
		as, _ := indexed[indexFn(orgName, subdomainId, envId)]
		return as
	}
	createAutogenSubdomain := func(orgName string, subdomainId string, envId string) types.AutogenSubdomain {
		as := types.AutogenSubdomain{
			IdModel:     types.IdModel{Id: 1},
			DnsName:     "xyz123",
			OrgName:     orgName,
			DomainName:  "nullstone.app",
			Fqdn:        "xyz123.nullstone.app.",
			Nameservers: []string{},
		}
		indexed[indexFn(orgName, subdomainId, envId)] = &as
		return as
	}

	router := mux.NewRouter()
	router.
		Methods(http.MethodPost).
		Path("/orgs/{orgName}/subdomains/{subdomainId}/envs/{envId}/autogen_subdomain").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainId, envId := vars["orgName"], vars["subdomainId"], vars["envId"]
			// NOTE: We're going to always return the same one we created instead of being random
			autogenSubdomain := createAutogenSubdomain(orgName, subdomainId, envId)
			raw, _ := json.Marshal(autogenSubdomain)
			w.Write(raw)
		})
	router.
		Methods(http.MethodDelete).
		Path("/orgs/{orgName}/subdomains/{subdomainId}/envs/{envId}/autogen_subdomain").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/subdomains/{subdomainId}/envs/{envId}/autogen_subdomain").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainId, envId := vars["orgName"], vars["subdomainId"], vars["envId"]
			autogenSubdomain := findAutogenSubdomain(orgName, subdomainId, envId)
			if autogenSubdomain != nil {
				raw, _ := json.Marshal(autogenSubdomain)
				w.Write(raw)
			} else {
				http.NotFound(w, r)
			}
		})
	router.
		Methods(http.MethodPut).
		Path("/orgs/{orgName}/subdomains/{subdomainId}/envs/{envId}/autogen_subdomain/delegation").
		Headers("Content-Type", "application/json").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainId, envId := vars["orgName"], vars["subdomainId"], vars["envId"]
			autogenSubdomain := findAutogenSubdomain(orgName, subdomainId, envId)
			if autogenSubdomain == nil {
				http.NotFound(w, r)
				return
			}

			if r.Body == nil {
				http.Error(w, "invalid body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()
			decoder := json.NewDecoder(r.Body)
			var delegation types.AutogenSubdomain
			if err := decoder.Decode(&delegation); err != nil {
				http.Error(w, fmt.Sprintf("invalid body: %s", err), http.StatusInternalServerError)
				return
			}

			autogenSubdomain.Nameservers = delegation.Nameservers
			raw, _ := json.Marshal(autogenSubdomain)
			w.Write(raw)
		})
	router.
		Methods(http.MethodDelete).
		Path("/orgs/{orgName}/subdomains/{subdomainId}/envs/{envId}/autogen_subdomain/delegation").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainId, envId := vars["orgName"], vars["subdomainId"], vars["envId"]
			if autogenSubdomain := findAutogenSubdomain(orgName, subdomainId, envId); autogenSubdomain == nil {
				http.NotFound(w, r)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})
	return router
}
