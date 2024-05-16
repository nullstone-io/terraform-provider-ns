package provider

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/http"
)

func mockNsServerWithAutogenSubdomains(autogenSubdomains map[string]map[string]map[string]*types.AutogenSubdomain) http.Handler {
	findAutogenSubdomain := func(orgName string, subdomainId string, envId string) *types.AutogenSubdomain {
		orgScoped, ok := autogenSubdomains[orgName]
		if !ok {
			return nil
		}
		subdomainScoped, ok := orgScoped[subdomainId]
		if !ok {
			return nil
		}
		result, ok := subdomainScoped[envId]
		if !ok {
			return nil
		}
		return result
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
		orgScoped, ok := autogenSubdomains[orgName]
		if !ok {
			orgScoped = map[string]map[string]*types.AutogenSubdomain{}
			autogenSubdomains[orgName] = orgScoped
		}
		subdomainScoped, ok := orgScoped[subdomainId]
		if !ok {
			subdomainScoped = map[string]*types.AutogenSubdomain{}
			orgScoped[subdomainId] = subdomainScoped
		}
		subdomainScoped[envId] = &as
		return as
	}

	router := mux.NewRouter()
	router.
		Methods(http.MethodPost).
		Path("/orgs/{orgName}/subdomains/{subdomainId}/envs/{envId}/autogen_subdomain").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgName := mux.Vars(r)["orgName"]
			// NOTE: We're going to always return the same one we created instead of being random
			autogenSubdomain := createAutogenSubdomain(orgName, "", "")
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
