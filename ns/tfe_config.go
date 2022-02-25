package ns

import (
	"github.com/hashicorp/go-tfe"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

func NewTfeConfig(apiConfig api.Config) *tfe.Config {
	cfg := tfe.DefaultConfig()
	if apiConfig.BaseAddress != "" {
		cfg.Address = apiConfig.BaseAddress
	}
	cfg.BasePath = "/terraform/v2/"
	// By default cfg.Token loads TFE_TOKEN env var
	// Fall back to TFE_TOKEN if api key is missing
	if apiConfig.ApiKey != "" {
		cfg.Token = apiConfig.ApiKey
	}
	return cfg
}
