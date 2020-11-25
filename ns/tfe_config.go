package ns

import (
	"github.com/hashicorp/go-tfe"
)

var (
	DefaultAddress = "https://api.nullstone.io"
)

func NewTfeConfig() *tfe.Config {
	cfg := tfe.DefaultConfig()
	if cfg.Address == "" {
		cfg.Address = DefaultAddress
	}
	cfg.BasePath = "/state/terraform/v2/"
	return cfg
}
