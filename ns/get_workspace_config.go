package ns

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func GetWorkspaceConfig(ctx context.Context, config api.Config, target types.WorkspaceTarget) (*types.RunConfig, error) {
	nsClient := api.Client{Config: config}
	workspace, err := nsClient.Workspaces().Get(ctx, target.StackId, target.BlockId, target.EnvId)
	if err != nil {
		return nil, err
	} else if workspace == nil {
		return nil, fmt.Errorf("no nullstone workspace %s", target.Id())
	}
	runConfig, err := nsClient.RunConfigs().GetLatest(ctx, workspace.StackId, workspace.Uid)
	if err != nil {
		return nil, err
	}
	return runConfig, nil
}
