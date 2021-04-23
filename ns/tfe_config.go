package ns

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"

	"github.com/hashicorp/go-tfe"
)

func NewTfeConfig() *tfe.Config {
	cfg := tfe.DefaultConfig()
	// Priority: NULLSTONE_ADDR > TFE_ADDRESS > DefaultAddress
	cfg.Address = os.Getenv(api.AddressEnvVar)
	if cfg.Address == "" {
		cfg.Address = os.Getenv("TFE_ADDRESS")
	}
	if cfg.Address == "" {
		cfg.Address = api.DefaultAddress
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
