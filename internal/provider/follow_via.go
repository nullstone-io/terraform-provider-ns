package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/workspaces"
	"log"
	"strings"
)

type ErrViaConnectionNotFound struct {
	Workspace types.WorkspaceTarget
	Via       string
}

// Is implements error wrapping Is(error) bool interface
func (e *ErrViaConnectionNotFound) Is(target error) bool {
	var vcnf *ErrViaConnectionNotFound
	return errors.As(target, &vcnf)
}

func (e *ErrViaConnectionNotFound) Error() string {
	return fmt.Sprintf("via connection (%s) was not found in workspace %s", e.Via, e.Workspace.Id())
}

// walkViaConnection traverses one or many connections to retrieve the target workspace and its connections
// If a via connection contains "/", it will use followViaConnection for each token separated by "/"
func walkViaConnection(ctx context.Context, nsConfig api.Config, sourceWorkspace types.WorkspaceTarget, connections types.Connections, localConnections workspaces.ManifestConnections, via string) (types.WorkspaceTarget, types.Connections, error) {
	curWorkspace, curConnections := sourceWorkspace, connections
	for _, via := range strings.Split(via, "/") {
		var err error
		curWorkspace, curConnections, err = followViaConnection(ctx, nsConfig, curWorkspace, curConnections, localConnections, via)
		if err != nil {
			return curWorkspace, curConnections, fmt.Errorf("error traversing via %q: %w", via, err)
		}
	}
	return curWorkspace, curConnections, nil
}

// followViaConnection traverses a single connection to retrieve the target workspace and its connections
func followViaConnection(ctx context.Context, nsConfig api.Config, sourceWorkspace types.WorkspaceTarget, connections types.Connections, localConnections workspaces.ManifestConnections, via string) (types.WorkspaceTarget, types.Connections, error) {
	viaWorkspace := findViaWorkspace(sourceWorkspace, connections, localConnections, via)
	if viaWorkspace == nil {
		return sourceWorkspace, connections, &ErrViaConnectionNotFound{Workspace: sourceWorkspace, Via: via}
	}

	log.Printf("(followViaConnection) Pulling (via=%s) connections for %s", via, viaWorkspace.Id())
	viaRunConfig, err := ns.GetWorkspaceConfig(ctx, nsConfig, *viaWorkspace)
	if err != nil {
		return sourceWorkspace, connections, fmt.Errorf("error retrieving connections for `via` workspace (via=%s, workspace=%s): %w", via, viaWorkspace.Id(), err)
	}
	return *viaWorkspace, viaRunConfig.Connections, nil
}

func findViaWorkspace(sourceWorkspace types.WorkspaceTarget, connections types.Connections, localConnections workspaces.ManifestConnections, via string) *types.WorkspaceTarget {
	// 1. Try local connections first
	mct, ok := localConnections[via]
	if ok {
		ct := &types.WorkspaceTarget{
			StackId: mct.StackId,
			BlockId: mct.BlockId,
			EnvId:   sourceWorkspace.EnvId,
		}
		if mct.EnvId != nil {
			ct.EnvId = *mct.EnvId
		}
		return ct
	}

	// 2. Try connections normally
	viaWorkspaceConn, ok := connections[via]
	if ok && viaWorkspaceConn.Reference != nil {
		ct := sourceWorkspace.FindRelativeConnection(*viaWorkspaceConn.Reference)
		return &ct
	}

	// 3. We can't find the workspace for the connection
	return nil
}
