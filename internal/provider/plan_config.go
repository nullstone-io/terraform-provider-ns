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
			OrgName:   os.Getenv("NULLSTONE_ORG"),
			StackName: os.Getenv("NULLSTONE_STACK"),
			EnvName:   os.Getenv("NULLSTONE_ENV"),
			BlockName: os.Getenv("NULLSTONE_BLOCK"),
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
