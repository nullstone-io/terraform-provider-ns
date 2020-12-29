package ns

import (
	"os"
	
	"github.com/hashicorp/go-tfe"
)

var (
	DefaultAddress = "https://api.nullstone.io"
)

func NewTfeConfig() *tfe.Config {
	cfg := tfe.DefaultConfig()
	cfg.Address = DefaultAddress
	if val := os.Getenv("TFE_ADDRESS"); val != "" {
		cfg.Address = val
	}
	cfg.BasePath = "/terraform/v2/"
	return cfg
}
