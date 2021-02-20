package provider

import (
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"net/http"
	"net/http/httptest"
)

func protoV5ProviderFactories(getNsConfig func() ns.Config, getTfeConfig func() *tfe.Config) map[string]func() (tfprotov5.ProviderServer, error) {
	return map[string]func() (tfprotov5.ProviderServer, error){
		"ns": func() (tfprotov5.ProviderServer, error) {
			return New("acctest", getNsConfig, getTfeConfig), nil
		},
	}
}

func mockTfe(handler http.Handler) (func() *tfe.Config, func()) {
	cfg := ns.NewTfeConfig()
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
