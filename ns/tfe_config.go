package ns

import (
	"github.com/hashicorp/go-tfe"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"
)

func NewTfeConfig(apiConfig api.Config) *tfe.Config {
	cfg := tfe.DefaultConfig()
	// Fall back to TFE_ADDRESS if api config does not have an address
	cfg.Address = os.Getenv(api.AddressEnvVar)
	if cfg.Address == "" {
		cfg.Address = os.Getenv("TFE_ADDRESS")
	}
	cfg.BasePath = "/terraform/v2/"
	// By default cfg.Token loads TFE_TOKEN env var
	// Fall back to TFE_TOKEN if api key is missing
	if apiConfig.ApiKey != "" {
		cfg.Token = apiConfig.ApiKey
	}
	return cfg
}
