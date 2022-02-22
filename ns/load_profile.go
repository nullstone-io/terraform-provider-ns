package ns

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"os"
)

const (
	NullstoneProfileEnvVar  = "NULLSTONE_PROFILE"
	DefaultNullstoneProfile = "default"
)

func LoadProfile() (*config.Profile, api.Config, error) {
	profileName := os.Getenv(NullstoneProfileEnvVar)
	if profileName == "" {
		profileName = DefaultNullstoneProfile
	}
	profile, err := config.LoadProfile(profileName)
	if err != nil {
		return nil, api.Config{}, err
	}

	cfg := api.DefaultConfig()
	if profile.Address != "" {
		cfg.BaseAddress = profile.Address
	}
	if profile.ApiKey != "" {
		cfg.ApiKey = profile.ApiKey
	}
	cfg.ApiKey = config.CleanseApiKey(cfg.ApiKey)

	// Load org name with the following precedence
	//  1. NULLSTONE_ORG env var
	//  2. ~/.nullstone/<profile>/org
	cfg.OrgName, _ = profile.LoadOrg()
	if val := os.Getenv("NULLSTONE_ORG"); val != "" {
		cfg.OrgName = val
	}

	return profile, cfg, nil
}
