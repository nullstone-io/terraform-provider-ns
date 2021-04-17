package ns

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"

	"github.com/hashicorp/go-tfe"
)

func NewTfeConfig() *tfe.Config {
	cfg := tfe.DefaultConfig()
	cfg.Address = api.DefaultAddress
	if val := os.Getenv(api.AddressEnvVar); val != "" {
		cfg.Address = val
	}
	cfg.BasePath = "/terraform/v2/"
	if val := os.Getenv(api.ApiKeyEnvVar); val != "" {
		cfg.Token = val
	}
	return cfg
}
