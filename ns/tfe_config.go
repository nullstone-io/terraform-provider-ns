package ns

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"

	"github.com/hashicorp/go-tfe"
)

func NewTfeConfig() *tfe.Config {
	cfg := tfe.DefaultConfig()
	// If TFE_ADDRESS is missing, we will look at NULLSTONE_ADDR, then use the default address
	if cfg.Address == "" {
		cfg.Address = api.DefaultAddress
		if val := os.Getenv(api.AddressEnvVar); val != "" {
			cfg.Address = val
		}
	}
	cfg.BasePath = "/terraform/v2/"
	// If TFE_TOKEN is missing, we will look at NULLSTONE_API_KEY
	if cfg.Token == "" {
		if val := os.Getenv(api.ApiKeyEnvVar); val != "" {
			cfg.Token = val
		}
	}
	return cfg
}
