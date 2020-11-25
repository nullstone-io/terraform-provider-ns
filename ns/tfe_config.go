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
	cfg.Token = os.Getenv("NULLSTONE_API_KEY")
	cfg.Address = os.Getenv("NULLSTONE_ADDR")
	if cfg.Address == "" {
		cfg.Address = DefaultAddress
	}
	cfg.BasePath = "/state/terraform/v2/"
	return cfg
}
