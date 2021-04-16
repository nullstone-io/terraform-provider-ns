package ns

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func GetWorkspaceConfig(config api.Config, target types.WorkspaceTarget) (*types.RunConfig, error) {
	nsClient := api.Client{Config: config}
	workspace, err := nsClient.Workspaces().Get(target.StackName, target.BlockName, target.EnvName)
	if err != nil {
		return nil, err
	} else if workspace == nil {
		return nil, fmt.Errorf("no nullstone workspace (stack=%s, env=%s, block=%s", target.StackName, target.EnvName, target.BlockName)
	}
	runConfig, err := nsClient.RunConfigs().GetLatest(workspace.StackName, workspace.Uid)
	if err != nil {
		return nil, err
	}
	return runConfig, nil
}
