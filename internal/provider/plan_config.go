package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type PlanConfig struct {
	Org         string            `json:"org"`
	Stack       string            `json:"stack"`
	Env         string            `json:"env"`
	Block       string            `json:"block"`
	Connections map[string]string `json:"connections"`
}

func (c PlanConfig) GetConnectionWorkspace(name string) string {
	if value, ok := c.Connections[name]; ok {
		return c.FullyQualifiedConnection(value)
	} else {
		value := os.Getenv(fmt.Sprintf(`NULLSTONE_CONNECTION_%s`, name))
		return c.FullyQualifiedConnection(value)
	}
}

func (c PlanConfig) FullyQualifiedConnection(name string) string {
	destStack, destEnv, destBlock := c.Stack, c.Env, ""

	tokens := strings.Split(name, ".")
	if len(tokens) > 2 {
		destStack = tokens[len(tokens)-3]
	}
	if len(tokens) > 1 {
		destEnv = tokens[len(tokens)-2]
	}
	destBlock = tokens[len(tokens)-1]

	return fmt.Sprintf("%s-%s-%s", destStack, destEnv, destBlock)
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
