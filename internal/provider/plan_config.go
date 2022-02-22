package provider

import (
	"encoding/json"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/workspaces"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strconv"
)

var (
	activeWorkspaceYmlFilename = path.Join(".nullstone", "active-workspace.yml")
	nullstoneJsonFilename      = ".nullstone.json"
)

type PlanConfig workspaces.Manifest

func (c PlanConfig) WorkspaceTarget() types.WorkspaceTarget {
	return types.WorkspaceTarget{
		StackId: c.StackId,
		BlockId: c.BlockId,
		EnvId:   c.EnvId,
	}
}

// LoadPlanConfig loads nullstone context for the current workspace
// Originally, this was in a file named `.nullstone.json`, but moved to `.nullstone/active-workspace.yml`
// As a result, this function will attempt the following:
//   1. Load `.nullstone/active-workspace.yml`
//   2. If not found, load `.nullstone.json`
//   3. Fall back to environment variables for each attribute
func LoadPlanConfig() (PlanConfig, error) {
	c := planConfigFromEnv()

	// Attempt to load .nullstone/active-workspace.yml
	if file, err := os.Open(activeWorkspaceYmlFilename); err == nil {
		decoder := yaml.NewDecoder(file)
		err2 := decoder.Decode(&c)
		file.Close()
		return c, err2
	}

	// Attempt to load .nullstone.json
	if file, err := os.Open(nullstoneJsonFilename); err == nil {
		decoder := json.NewDecoder(file)
		err2 := decoder.Decode(&c)
		file.Close()
		return c, err2
	}

	// Just rely on config from env if no plan config files
	return c, nil
}

func planConfigFromEnv() PlanConfig {
	return PlanConfig{
		OrgName: os.Getenv("NULLSTONE_ORG_NAME"),

		StackId:   readIntFromEnvVars("NULLSTONE_STACK_ID"),
		StackName: os.Getenv("NULLSTONE_STACK_NAME"),

		BlockId:   readIntFromEnvVars("NULLSTONE_BLOCK_ID"),
		BlockName: os.Getenv("NULLSTONE_BLOCK_NAME"),
		BlockRef:  os.Getenv("NULLSTONE_BLOCK_REF"),

		EnvId:   readIntFromEnvVars("NULLSTONE_ENV_ID"),
		EnvName: os.Getenv("NULLSTONE_ENV_NAME"),
	}
}

func readIntFromEnvVars(name string) int64 {
	if val, err := strconv.ParseInt(os.Getenv(name), 10, 64); err == nil {
		return val
	}
	return 0
}
