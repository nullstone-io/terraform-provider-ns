package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
	"regexp"
)

var validConnectionName = regexp.MustCompile("^[_a-z0-9/-]+$")

type dataConnection struct {
	p *provider
}

func newDataConnection(p *provider) (*dataConnection, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataConnection{p: p}, nil
}

func (*dataConnection) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to configure a connection to another nullstone workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "name",
					Type:            tftypes.String,
					Required:        true,
					Description:     "The unique name of the connection within this module.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "type",
					Type:            tftypes.String,
					Required:        true,
					Description:     "The type of module to satisfy this connection.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "optional",
					Type:            tftypes.Bool,
					Optional:        true,
					Description:     "This data source will cause an error if optional is false and this connection is not configured.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:     "via",
					Type:     tftypes.String,
					Optional: true,
					Description: `Defines this connection is satisfied through another ns_connection.
Typically, this is set to data.ns_connection.other.name`,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "workspace_id",
					Type:            tftypes.String,
					Computed:        true,
					Description:     "This refers to the workspace in nullstone. This follows the form `{stack}/{env}/{block}`.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "outputs",
					Type:            tftypes.DynamicPseudoType,
					Computed:        true,
					Description:     `An object containing every root-level output in the remote state.`,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (d *dataConnection) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataConnection) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	nsClient := api.Client{Config: d.p.NsConfig}

	name := extractStringFromConfig(config, "name")
	type_ := extractStringFromConfig(config, "type")
	optional := extractBoolFromConfig(config, "optional")
	via := extractStringFromConfig(config, "via")
	workspaceId := ""

	diags := make([]*tfprotov5.Diagnostic, 0)
	if !validConnectionName.Match([]byte(name)) {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("name (%s) can only contain the characters 'a'-'z', '0'-'9', '-', '_'", name),
		})
	}
	if len(diags) > 0 {
		return nil, diags, nil
	}

	outputsValue := tftypes.NewValue(tftypes.Map{AttributeType: tftypes.String}, map[string]tftypes.Value{})

	workspace, err := d.getConnectionWorkspace(name, type_, via)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to find nullstone workspace.",
			Detail:   err.Error(),
		})
	} else if workspace != nil {
		workspaceId = workspace.Id()
		nfWorkspace, err := nsClient.Workspaces().Get(workspace.StackId, workspace.BlockId, workspace.EnvId)
		if err != nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf(`Unable to find nullstone workspace %s`, workspace.Id()),
				Detail:   err.Error(),
			})
		} else {
			stateFile, err := ns.GetStateFile(d.p.TfeClient, d.p.PlanConfig.OrgName, nfWorkspace.Uid.String())
			if err != nil {
				diags = append(diags, &tfprotov5.Diagnostic{
					Severity: tfprotov5.DiagnosticSeverityWarning,
					Summary:  fmt.Sprintf(`Unable to download workspace outputs for %q. 'outputs' will be empty`, workspace),
					Detail:   err.Error(),
				})
			} else {
				if ov, err := stateFile.Outputs.ToProtov5(); err != nil {
					diags = append(diags, &tfprotov5.Diagnostic{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  fmt.Sprintf(`Unable to read workspace outputs for %q. 'outputs' will be empty`, workspace),
						Detail:   err.Error(),
					})
				} else {
					outputsValue = ov
				}
			}
		}
	} else if !optional {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The connection %q is missing. It is required to use this plan.", name),
		})
	}

	return map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, fmt.Sprintf("%s-%s", name, workspaceId)),
		"name":         tftypes.NewValue(tftypes.String, name),
		"type":         tftypes.NewValue(tftypes.String, type_),
		"workspace_id": tftypes.NewValue(tftypes.String, workspaceId),
		"optional":     tftypes.NewValue(tftypes.Bool, optional),
		"via":          tftypes.NewValue(tftypes.String, via),
		"outputs":      outputsValue,
	}, diags, nil
}

func (d *dataConnection) getConnectionWorkspace(name, type_, via string) (*types.WorkspaceTarget, error) {
	sourceWorkspace := d.p.PlanConfig.WorkspaceTarget()

	log.Printf("(getConnectionWorkspace) Pulling workspace run config for @ %s", sourceWorkspace.Id())
	runConfig, err := ns.GetWorkspaceConfig(d.p.NsConfig, sourceWorkspace)
	if err != nil {
		return nil, err
	}

	connections := runConfig.Connections
	capabilityId := d.p.PlanConfig.CapabilityId
	if capabilityId > 0 {
		for _, cap := range runConfig.Capabilities {
			if cap.Id == capabilityId {
				connections = cap.Connections
				break
			}
		}
	}

	// If this data_connection has `via` specified, then we need to
	//   get the connections for *that* workspace instead of the current workspace
	if via != "" {
		viaWorkspaceConn, ok := connections[via]
		if !ok || viaWorkspaceConn.Reference == nil {
			log.Printf("via connection (%s) was not found in %s", via, sourceWorkspace.Id())
			return nil, nil
		}
		viaWorkspace := sourceWorkspace.FindRelativeConnection(*viaWorkspaceConn.Reference)
		log.Printf("(getConnectionWorkspace) Pulling (via=%s) connections for %s", via, viaWorkspace.Id())
		viaRunConfig, err := ns.GetWorkspaceConfig(d.p.NsConfig, viaWorkspace)
		if err != nil {
			return nil, fmt.Errorf("error retrieving connections for `via` workspace (via=%s, workspace=%s): %w", via, viaWorkspace.Id(), err)
		}
		sourceWorkspace = viaWorkspace
		runConfig = viaRunConfig
		connections = runConfig.Connections
	}

	conn, ok := connections[name]
	if !ok || conn.Reference == nil {
		log.Printf("connection (%s) was not found in %s", name, sourceWorkspace.Id())
		return nil, nil
	}
	if conn.Type != type_ {
		return nil, fmt.Errorf("retrieved connection, but the connection types do not match (desired=%s, actual=%s)", type_, conn.Type)
	}
	found := sourceWorkspace.FindRelativeConnection(*conn.Reference)
	log.Printf("(getConnectionWorkspace) Found workspace in connections @ %s", found.Id())
	return &found, nil
}
