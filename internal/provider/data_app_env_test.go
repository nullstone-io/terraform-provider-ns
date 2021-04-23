package provider

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/http"
	"strconv"
	"testing"
)

func TestDataAppEnv(t *testing.T) {
	core := &types.Stack{
		IdModel:     types.IdModel{
			Id: 2,
		},
		Name:        "core",
		OrgName:     "org0",
	}
	app1 := &types.Application{
		IdModel: types.IdModel{
			Id: 1,
		},
		Name:      "app1",
		OrgName:   "org0",
		StackName: core.Name,
	}
	dev := &types.Environment{
		IdModel: types.IdModel{
			Id: 1,
		},
		Name:      "dev",
		OrgName:   "org0",
		StackName: core.Name,
	}
	prod := &types.Environment{
		IdModel: types.IdModel{
			Id: 2,
		},
		Name:      "prod",
		OrgName:   "org0",
		StackName: core.Name,
	}

	appEnvs := []*types.AppEnv{
		{
			IdModel: types.IdModel{
				Id: 5,
			},
			AppId:   app1.Id,
			EnvId:   dev.Id,
			Version: "1.0.0",
			App:     app1,
			Env:     prod,
		},
	}
	apps := []*types.Application{app1}
	envs := []*types.Environment{dev,prod}

	t.Run("sets up attributes properly with new AppEnv", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}

data "ns_workspace" "this" {
  stack = "core"
  block = "app1"
  env   = "dev"
}

data "ns_app_env" "this" {
  app   = data.ns_workspace.this.block
  stack = data.ns_workspace.this.stack
  env   = data.ns_workspace.this.env
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_app_env.this", `id`, "10"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `app`, "app1"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `stack`, "core"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `env`, "dev"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `version`, ""),
		)

		getNsConfig, closeNsFn := mockNs(mockNsHandlerAppEnvs(&appEnvs, apps, envs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(nil)
		defer closeTfeFn()

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig),
			Steps: []resource.TestStep{
				{
					Config: tfconfig,
					Check:  checks,
				},
			},
		})
	})

	t.Run("sets up attributes properly with existing AppEnv", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}

data "ns_workspace" "this" {
  stack = "core"
  block = "app1"
  env   = "prod"
}

data "ns_app_env" "this" {
  app   = data.ns_workspace.this.block
  stack = data.ns_workspace.this.stack
  env   = data.ns_workspace.this.env
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_app_env.this", `id`, "5"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `app`, "app1"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `stack`, "core"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `env`, "prod"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `version`, "1.0.0"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsHandlerAppEnvs(&appEnvs, apps, envs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(nil)
		defer closeTfeFn()

		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: protoV5ProviderFactories(getNsConfig, getTfeConfig),
			Steps: []resource.TestStep{
				{
					Config: tfconfig,
					Check:  checks,
				},
			},
		})
	})
}

func mockNsHandlerAppEnvs(appEnvs *[]*types.AppEnv, apps []*types.Application, envs []*types.Environment) http.Handler {
	findApp := func(orgName, appIdStr string) *types.Application {
		for _, app := range apps {
			if app.OrgName == orgName && strconv.Itoa(app.Id) == appIdStr {
				return app
			}
		}
		return nil
	}
	findEnv := func(orgName, envName string) *types.Environment {
		for _, env := range envs {
			if env.OrgName == orgName && env.Name == envName {
				return env
			}
		}
		return nil
	}
	getAppEnv := func(orgName, appIdStr, envName string) *types.AppEnv {
		for _, existing := range *appEnvs {
			if existing.App.OrgName == orgName && strconv.Itoa(existing.App.Id) == appIdStr &&
				existing.Env.OrgName == orgName && existing.Env.Name == envName {
				return existing
			}
		}
		return nil
	}
	addAppEnv := func(orgName, appIdStr, envName string) *types.AppEnv {
		app := findApp(orgName, appIdStr)
		env := findEnv(orgName, envName)
		if app == nil || env == nil {
			return nil
		}

		appEnv := &types.AppEnv{
			IdModel: types.IdModel{
				Id: 10,
			},
			AppId:   app.Id,
			EnvId:   env.Id,
			Version: "",
			App:     app,
			Env:     env,
		}
		*appEnvs = append(*appEnvs, appEnv)
		return appEnv
	}
	findOrCreateEnv := func(r *http.Request) *types.AppEnv {
		vars := mux.Vars(r)
		orgName, appId, envName := vars["orgName"], vars["appId"], vars["envName"]

		appEnv := getAppEnv(orgName, appId, envName)
		if appEnv == nil {
			appEnv = addAppEnv(orgName, appId, envName)
		}
		return appEnv
	}

	router := mux.NewRouter()
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/apps").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := json.Marshal(apps)
			w.Write(raw)
		})
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/apps/{appId}/envs/{envName}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appEnv := findOrCreateEnv(r); appEnv == nil {
				http.NotFound(w, r)
			} else {
				raw, _ := json.Marshal(appEnv)
				w.Write(raw)
			}
		})
	router.
		Methods(http.MethodPut).
		Path("/orgs/{orgName}/apps/{appId}/envs/{envName}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := struct{
				Version string `json:"version"`
			}{}
			decoder := json.NewDecoder(r.Body)
			if appEnv := findOrCreateEnv(r); appEnv == nil {
				http.NotFound(w, r)
			} else if err := decoder.Decode(&payload); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			} else {
				appEnv.Version = payload.Version
				raw, _ := json.Marshal(appEnv)
				w.Write(raw)
			}
		})
	return router
}
