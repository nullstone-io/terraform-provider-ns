package ns

import (
	"os"

	"github.com/hashicorp/go-tfe"
)

var (
	DefaultTfeAddress = "https://api.nullstone.io"
)

func NewTfeConfig() *tfe.Config {
	cfg := tfe.DefaultConfig()
	cfg.Address = DefaultTfeAddress
	if val := os.Getenv("TFE_ADDRESS"); val != "" {
		cfg.Address = val
	}
	cfg.BasePath = "/terraform/v2/"
	return cfg
}
