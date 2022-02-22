package provider

import (
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"net/http"
	"net/http/httptest"
)

func protoV5ProviderFactories(getNsConfig func() api.Config, getTfeConfig func() *tfe.Config) map[string]func() (tfprotov5.ProviderServer, error) {
	return map[string]func() (tfprotov5.ProviderServer, error){
		"ns": func() (tfprotov5.ProviderServer, error) {
			return Mock("acctest", getNsConfig, getTfeConfig), nil
		},
	}
}

func mockNs(handler http.Handler) (func() api.Config, func()) {
	cfg := api.DefaultConfig()
	cfg.ApiKey = "abcdefgh012345789"
	fn := func() api.Config {
		return cfg
	}
	if handler == nil {
		return fn, func() {}
	}

	server := httptest.NewServer(handler)
	cfg.BaseAddress = server.URL
	return fn, server.Close
}

func mockTfe(handler http.Handler) (func() *tfe.Config, func()) {
	cfg := ns.NewTfeConfig(api.Config{})
	cfg.Token = "abcdefgh012345789"
	fn := func() *tfe.Config {
		return cfg
	}
	if handler == nil {
		return fn, func() {}
	}

	server := httptest.NewServer(handler)
	cfg.Address = server.URL
	return fn, server.Close
}
