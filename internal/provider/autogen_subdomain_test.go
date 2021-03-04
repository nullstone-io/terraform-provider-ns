package provider

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"net/http"
)

func mockNsServerWithAutogenSubdomains(subdomains map[string]map[string]ns.AutogenSubdomain, delegations map[string]map[string]ns.AutogenSubdomainDelegation) http.Handler {
	findSubdomain := func(orgName, subdomainName string) *ns.AutogenSubdomain {
		orgSubdomains, ok := subdomains[orgName]
		if !ok {
			return nil
		}
		subdomain, ok := orgSubdomains[subdomainName]
		if !ok {
			return nil
		}
		return &subdomain
	}
	findDelegation := func(orgName, subdomainName string) *ns.AutogenSubdomainDelegation {
		orgDelegations, ok := delegations[orgName]
		if !ok {
			return nil
		}
		delegation, ok := orgDelegations[subdomainName]
		if !ok {
			return nil
		}
		return &delegation
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
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			orgName, subdomainName := vars["orgName"], vars["subdomainName"]
			if _, ok := delegations[orgName]; !ok {
				delegations[orgName] = map[string]ns.AutogenSubdomainDelegation{}
			}

			if r.Body == nil {
				http.Error(w, "invalid body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()
			decoder := json.NewDecoder(r.Body)
			var nameservers []string
			if err := decoder.Decode(&nameservers); err != nil {
				http.Error(w, "invalid body", http.StatusInternalServerError)
				return
			}

			delegation := ns.AutogenSubdomainDelegation{
				Nameservers: nameservers,
			}
			delegations[orgName][subdomainName] = delegation
			raw, _ := json.Marshal(delegation)
			w.Write(raw)
		})
	return router
}
