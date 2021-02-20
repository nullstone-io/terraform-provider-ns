package provider

import (
	"encoding/json"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"io/ioutil"
	"os"
)

type PlanConfig struct {
	ns.WorkspaceLocation
	Org string `json:"org"`
}

func PlanConfigFromEnv() PlanConfig {
	return PlanConfig{
		WorkspaceLocation: ns.WorkspaceLocationFromEnv(),
		Org:               os.Getenv("NULLSTONE_ORG"),
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
