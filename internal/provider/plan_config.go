package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type PlanConfig struct {
	Org         string            `json:"org"`
	Stack       string            `json:"stack"`
	Env         string            `json:"env"`
	Block       string            `json:"block"`
	Connections map[string]string `json:"connections"`
}

func (c PlanConfig) GetConnection(name string) string {
	if value, ok := c.Connections[name]; ok {
		return value
	}
	return os.Getenv(fmt.Sprintf(`NULLSTONE_CONNECTION_%s`, name))
}

func DefaultPlanConfig() PlanConfig {
	return PlanConfig{
		Org:         os.Getenv("NULLSTONE_ORG"),
		Stack:       os.Getenv("NULLSTONE_STACK"),
		Env:         os.Getenv("NULLSTONE_ENV"),
		Block:       os.Getenv("NULLSTONE_BLOCK"),
		Connections: map[string]string{},
	}
}

func PlanConfigFromFile(filename string) (PlanConfig, error) {
	c := DefaultPlanConfig()
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, err
	}
	return c, nil
}
