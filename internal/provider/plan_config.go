package provider

import (
	"encoding/json"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"io/ioutil"
	"os"
	"strconv"
)

type PlanConfig struct {
	OrgName string `json:"orgName"`

	StackId   int64  `json:"stackId"`
	StackName string `json:"stackName"`

	BlockId   int64  `json:"blockId"`
	BlockName string `json:"blockName"`
	BlockRef  string `json:"blockRef"`

	EnvId   int64  `json:"envId"`
	EnvName string `json:"envName"`
}

func (c PlanConfig) WorkspaceTarget() types.WorkspaceTarget {
	return types.WorkspaceTarget{
		StackId: c.StackId,
		BlockId: c.BlockId,
		EnvId:   c.EnvId,
	}
}

func PlanConfigFromEnv() PlanConfig {
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

func PlanConfigFromFile(filename string) (PlanConfig, error) {
	c := PlanConfigFromEnv()
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, err
	}
	return c, nil
}
