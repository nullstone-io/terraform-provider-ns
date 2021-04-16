package ns

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-tfe"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
)

type StateFile struct {
	Version          int     `json:"version"`
	TerraformVersion string  `json:"terraform_version"`
	Serial           int64   `json:"serial"`
	Lineage          string  `json:"lineage"`
	Outputs          Outputs `json:"outputs"`
}

func GetStateFile(tfeClient *tfe.Client, orgName string, target types.WorkspaceTarget) (*StateFile, error) {
	workspaceName := fmt.Sprintf("%s-%s-%s", target.StackName, target.EnvName, target.BlockName)

	log.Printf("[DEBUG] Retrieving state file (org=%s, workspace=%s)\n", orgName, workspaceName)

	workspace, err := tfeClient.Workspaces.Read(context.Background(), orgName, workspaceName)
	if err != nil {
		return nil, fmt.Errorf(`error reading workspace (org=%s, workspace=%s): %w`, orgName, workspaceName, err)
	}
	log.Printf("[DEBUG] Found workspace (org=%s, workspace=%s), workspace id=%s", orgName, workspaceName, workspace.ID)

	sv, err := tfeClient.StateVersions.Current(context.Background(), workspace.ID)
	if err != nil {
		return nil, fmt.Errorf(`error reading current state version (org=%s, workspace=%s): %w`, orgName, workspaceName, err)
	}

	log.Printf("[DEBUG] Downloading state file (org=%s, workspace=%s) from %s", orgName, workspaceName, sv.DownloadURL)
	state, err := tfeClient.StateVersions.Download(context.Background(), sv.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf(`error downloading state file (org=%s, workspace=%s): %w`, orgName, workspaceName, err)
	}

	log.Printf("[DEBUG] Retrieved state file (org=%s, workspace=%s): size=%d\n", orgName, workspaceName, len(state))

	var stateFile StateFile
	if err := json.Unmarshal(state, &stateFile); err != nil {
		return nil, fmt.Errorf(`error parsing state file (org=%s, workspace=%s): %w`, orgName, workspaceName, err)
	}
	return &stateFile, nil
}
