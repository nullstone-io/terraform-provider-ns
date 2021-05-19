package provider

import (
	"encoding/json"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"io/ioutil"
	"os"
)

type PlanConfig struct {
	types.WorkspaceTarget
}

func PlanConfigFromEnv() PlanConfig {
	return PlanConfig{
		WorkspaceTarget: types.WorkspaceTarget{
			OrgName:   os.Getenv("NULLSTONE_ORG_NAME"),
			StackName: os.Getenv("NULLSTONE_STACK_NAME"),
			EnvName:   os.Getenv("NULLSTONE_ENV_NAME"),
			BlockName: os.Getenv("NULLSTONE_BLOCK_NAME"),
		},
	}
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
