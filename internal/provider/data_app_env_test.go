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
		IdModel: types.IdModel{
			Id: 2,
		},
		Name:    "core",
		OrgName: "org0",
	}
	app1 := &types.Application{
		Block: types.Block{
			IdModel: types.IdModel{
				Id: 1,
			},
			OrgName:   "org0",
			StackId:   core.Id,
			StackName: core.Name,
			Name:      "app1",
			Reference: "yellow-giraffe",
		},
	}
	dev := &types.Environment{
		IdModel: types.IdModel{
			Id: 1,
		},
		Name:      "dev",
		OrgName:   "org0",
		StackId:   core.Id,
		StackName: core.Name,
	}
	prod := &types.Environment{
		IdModel: types.IdModel{
			Id: 2,
		},
		Name:      "prod",
		OrgName:   "org0",
		StackId:   core.Id,
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
	envs := []*types.Environment{dev, prod}

	t.Run("sets up attributes properly with new AppEnv", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}

data "ns_workspace" "this" {
  stack_id   = 2
  stack_name = "stack0"
  block_id   = 1
  block_name = "app1"
  block_ref  = "yellow-giraffe"
  env_id     = 1
  env_name   = "dev"
}

data "ns_app_env" "this" {
  stack_id = data.ns_workspace.this.stack_id
  app_id   = data.ns_workspace.this.block_id
  env_id   = data.ns_workspace.this.env_id
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_app_env.this", `id`, "10"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `stack_id`, "2"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `app_id`, "1"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `env_id`, "1"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `version`, ""),
		)

		getNsConfig, closeNsFn := mockNs(mockNsHandlerAppEnvs(&appEnvs, apps, envs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(nil)
		defer closeTfeFn()

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

	t.Run("sets up attributes properly with existing AppEnv", func(t *testing.T) {
		tfconfig := fmt.Sprintf(`
provider "ns" {
  organization = "org0"
}

data "ns_workspace" "this" {
  stack_id   = 2
  stack_name = "stack0"
  block_id   = 1
  block_name = "app1"
  block_ref  = "yellow-giraffe"
  env_id     = 2
  env_name   = "prod"
}

data "ns_app_env" "this" {
  stack_id = data.ns_workspace.this.stack_id
  app_id   = data.ns_workspace.this.block_id
  env_id   = data.ns_workspace.this.env_id
}
`)
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.ns_app_env.this", `id`, "5"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `stack_id`, "2"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `app_id`, "1"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `env_id`, "2"),
			resource.TestCheckResourceAttr("data.ns_app_env.this", `version`, "1.0.0"),
		)

		getNsConfig, closeNsFn := mockNs(mockNsHandlerAppEnvs(&appEnvs, apps, envs))
		defer closeNsFn()
		getTfeConfig, closeTfeFn := mockTfe(nil)
		defer closeTfeFn()

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

func mockNsHandlerAppEnvs(appEnvs *[]*types.AppEnv, apps []*types.Application, envs []*types.Environment) http.Handler {
	findApp := func(orgName, appIdStr string) *types.Application {
		for _, app := range apps {
			if app.OrgName == orgName && strconv.FormatInt(app.Id, 10) == appIdStr {
				return app
			}
		}
		return nil
	}
	findEnvByName := func(orgName, envName string) *types.Environment {
		for _, env := range envs {
			if env.OrgName == orgName && env.Name == envName {
				return env
			}
		}
		return nil
	}
	findEnv := func(orgName, stackId, envId string) *types.Environment {
		for _, env := range envs {
			if env.OrgName == orgName && fmt.Sprintf("%d", env.StackId) == stackId && fmt.Sprintf("%d", env.Id) == envId {
				return env
			}
		}
		return nil
	}
	getAppEnv := func(orgName, appIdStr, envName string) *types.AppEnv {
		for _, existing := range *appEnvs {
			if existing.App.OrgName == orgName && fmt.Sprintf("%d", existing.App.Id) == appIdStr &&
				existing.Env.OrgName == orgName && existing.Env.Name == envName {
				return existing
			}
		}
		return nil
	}
	addAppEnv := func(orgName, appIdStr, envName string) *types.AppEnv {
		app := findApp(orgName, appIdStr)
		env := findEnvByName(orgName, envName)
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
		Path("/orgs/{orgName}/stacks/{stackId}/apps/{appId}/envs/{envName}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appEnv := findOrCreateEnv(r); appEnv == nil {
				http.NotFound(w, r)
			} else {
				raw, _ := json.Marshal(appEnv)
				w.Write(raw)
			}
		})
	router.
		Methods(http.MethodGet).
		Path("/orgs/{orgName}/stacks/{stackId}/envs/{envId}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			if env := findEnv(vars["orgName"], vars["stackId"], vars["envId"]); env == nil {
				http.NotFound(w, r)
			} else {
				raw, _ := json.Marshal(env)
				w.Write(raw)
			}
		})
	router.
		Methods(http.MethodPut).
		Path("/orgs/{orgName}/apps/{appId}/envs/{envName}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := struct {
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
