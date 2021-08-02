package provider

import (
	"errors"
	"fmt"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
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

func followViaConnection(nsConfig api.Config, sourceWorkspace types.WorkspaceTarget, connections types.Connections, via string) (types.WorkspaceTarget, types.Connections, error) {
	viaWorkspaceConn, ok := connections[via]
	if !ok || viaWorkspaceConn.Reference == nil {
		return sourceWorkspace, connections, &ErrViaConnectionNotFound{Workspace: sourceWorkspace, Via: via}
	}
	viaWorkspace := sourceWorkspace.FindRelativeConnection(*viaWorkspaceConn.Reference)
	log.Printf("(followViaConnection) Pulling (via=%s) connections for %s", via, viaWorkspace.Id())
	viaRunConfig, err := ns.GetWorkspaceConfig(nsConfig, viaWorkspace)
	if err != nil {
		return sourceWorkspace, connections, fmt.Errorf("error retrieving connections for `via` workspace (via=%s, workspace=%s): %w", via, viaWorkspace.Id(), err)
	}
	return viaWorkspace, viaRunConfig.Connections, nil
}
