package provider

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/http"
)

func mockNsServerWithAutogenSubdomains(subdomains map[string]map[string]*types.AutogenSubdomain, delegations map[string]map[string]*types.AutogenSubdomainDelegation) http.Handler {
	findSubdomain := func(orgName, subdomainName string) *types.AutogenSubdomain {
		orgSubdomains, ok := subdomains[orgName]
		if !ok {
			return nil
		}
		subdomain, ok := orgSubdomains[subdomainName]
		if !ok {
			return nil
		}
		return subdomain
	}
	findDelegation := func(orgName, subdomainName string) *types.AutogenSubdomainDelegation {
		orgDelegations, ok := delegations[orgName]
		if !ok {
			return nil
		}
		delegation, ok := orgDelegations[subdomainName]
		if !ok {
			return nil
		}
		return delegation
	}

	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/autogen_subdomains/{subdomainName}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainName := vars["orgName"], vars["subdomainName"]
			subdomain := findSubdomain(orgName, subdomainName)
			if subdomain != nil {
				raw, _ := json.Marshal(subdomain)
				w.Write(raw)
			} else {
				http.NotFound(w, r)
			}
		})
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/autogen_subdomains/{subdomainName}/delegation").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainName := vars["orgName"], vars["subdomainName"]
			delegation := findDelegation(orgName, subdomainName)
			if delegation != nil {
				raw, _ := json.Marshal(delegation)
				w.Write(raw)
				return
			} else {
				http.NotFound(w, r)
			}
		})
	router.
		Methods(http.MethodPut).
		Path("/orgs/{orgName}/autogen_subdomains/{subdomainName}/delegation").
		Headers("Content-Type", "application/json").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainName := vars["orgName"], vars["subdomainName"]
			if subdomain := findSubdomain(orgName, subdomainName); subdomain == nil {
				http.NotFound(w, r)
				return
			}
			if _, ok := delegations[orgName]; !ok {
				delegations[orgName] = map[string]*types.AutogenSubdomainDelegation{}
			}

			if r.Body == nil {
				http.Error(w, "invalid body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()
			decoder := json.NewDecoder(r.Body)
			var delegation types.AutogenSubdomainDelegation
			if err := decoder.Decode(&delegation); err != nil {
				http.Error(w, fmt.Sprintf("invalid body: %s", err), http.StatusInternalServerError)
				return
			}

			delegations[orgName][subdomainName] = &delegation
			raw, _ := json.Marshal(delegation)
			w.Write(raw)
		})
	router.
		Methods(http.MethodDelete).
		Path("/orgs/{orgName}/autogen_subdomains/{subdomainName}/delegation").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainName := vars["orgName"], vars["subdomainName"]
			if subdomain := findSubdomain(orgName, subdomainName); subdomain == nil {
				http.NotFound(w, r)
				return
			}
			if _, ok := delegations[orgName]; !ok {
				delegations[orgName] = map[string]*types.AutogenSubdomainDelegation{}
			}

			delegations[orgName][subdomainName] = &types.AutogenSubdomainDelegation{Nameservers: []string{}}
			w.WriteHeader(http.StatusNoContent)
		})
	return router
}
