package ns

import (
	"fmt"
	"os"
	"strings"
)

type WorkspaceLocation struct {
	Stack string `json:"stack"`
	Env   string `json:"env"`
	Block string `json:"block"`
}

func (w WorkspaceLocation) Id() string {
	return fmt.Sprintf("%s/%s/%s", w.Stack, w.Env, w.Block)
}

func WorkspaceLocationFromEnv() WorkspaceLocation {
	return WorkspaceLocation{
		Stack: os.Getenv("NULLSTONE_STACK"),
		Env:   os.Getenv("NULLSTONE_ENV"),
		Block: os.Getenv("NULLSTONE_BLOCK"),
	}
}

func FullyQualifiedWorkspace(baseStack, baseEnv, target string) *WorkspaceLocation {
	dest := &WorkspaceLocation{
		Stack: baseStack,
		Env:   baseEnv,
		Block: "",
	}

	tokens := strings.Split(target, ".")
	if len(tokens) > 2 {
		dest.Stack = tokens[len(tokens)-3]
	}
	if len(tokens) > 1 {
		dest.Env = tokens[len(tokens)-2]
	}
	dest.Block = tokens[len(tokens)-1]

	return dest
}
