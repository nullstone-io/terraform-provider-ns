package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var validConnectionName = regexp.MustCompile("^[_a-z0-9/-]+$")

var _ server.DataSource = &dataConnection{}

type dataConnection struct {
	p               *provider
	isAppConnection bool
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
					Optional:        true,
					Description:     "The type of module to satisfy this connection.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "contract",
					Type:            tftypes.String,
					Optional:        true,
					Description:     "The contract that defines which modules can satisfy this connection.",
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
	contract := extractStringFromConfig(config, "contract")
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
	if type_ == "" && contract == "" {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("contract is required"),
		})
	}
	contractName, err := types.ParseModuleContractName(contract)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("contract (%s) is invalid: %s", contract, err),
		})
	}
	if len(diags) > 0 {
		return nil, diags, nil
	}

	outputsValue := tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{})

	workspace, err := d.getConnectionWorkspace(ctx, name, contractName, type_, via)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to find nullstone workspace.",
			Detail:   err.Error(),
		})
	} else if workspace != nil {
		workspaceId = workspace.Id()
		nfWorkspace, err := nsClient.Workspaces().Get(ctx, workspace.StackId, workspace.BlockId, workspace.EnvId)
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
					Summary:  fmt.Sprintf(`Unable to download workspace outputs for %q. 'outputs' will be empty`, workspace.Id()),
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
		"contract":     tftypes.NewValue(tftypes.String, contract),
		"workspace_id": tftypes.NewValue(tftypes.String, workspaceId),
		"optional":     tftypes.NewValue(tftypes.Bool, optional),
		"via":          tftypes.NewValue(tftypes.String, via),
		"outputs":      outputsValue,
	}, diags, nil
}

func (d *dataConnection) getConnectionWorkspace(ctx context.Context, name string, contractName types.ModuleContractName, type_, via string) (*types.WorkspaceTarget, error) {
	log.Printf("(getConnectionWorkspace) name=%s contract=%s type=%s via=%s capabilityName=%d", name, contractName, type_, via, d.p.PlanConfig.CapabilityName)
	sourceWorkspace := d.p.PlanConfig.WorkspaceTarget()

	// Let's search for a configured connection in .nullstone/active-workspace.yml first
	localConnections := d.p.PlanConfig.Connections
	if localConnections != nil {
		if reference, ok := localConnections[name]; ok {
			ct := types.ConnectionTarget{
				StackId:   reference.StackId,
				BlockId:   reference.BlockId,
				BlockName: reference.BlockName,
				EnvId:     reference.EnvId,
			}
			found := sourceWorkspace.FindRelativeConnection(ct)
			log.Printf("(getConnectionWorkspace) Found workspace defined in plan config @ %s", found.Id())
			return &found, nil
		}
	}

	log.Printf("(getConnectionWorkspace) Pulling workspace run config for @ %s", sourceWorkspace.Id())
	runConfig, err := ns.GetWorkspaceConfig(ctx, d.p.NsConfig, sourceWorkspace)
	if err != nil {
		return nil, err
	}

	// If this data_connection is established on the capability, we need to pull from the correct set of connections
	connections := d.getConnectionsFromRunConfig(runConfig)
	raw, _ := json.Marshal(connections)
	log.Printf("(getConnectionWorkspace) Utilizing connections (capability name=%d) %s", d.p.PlanConfig.CapabilityName, string(raw))

	// If this data_connection has `via` specified, then we need to
	//   get the connections for *that* workspace instead of the current workspace
	if via != "" {
		sourceWorkspace, connections, err = walkViaConnection(ctx, d.p.NsConfig, sourceWorkspace, connections, localConnections, via)
		if errors.Is(err, &ErrViaConnectionNotFound{}) {
			log.Printf("(getConnectionWorkspace) %s\n", err)
			return nil, nil
		} else if err != nil {
			return nil, err
		}
	}

	conn, ok := connections[name]
	if !ok || conn.EffectiveTarget == nil {
		log.Printf("(getConnectionWorkspace) Connection (%s) was not found in %s", name, sourceWorkspace.Id())
		return nil, nil
	}
	if err := d.validateConnection(conn, contractName, type_); err != nil {
		return nil, fmt.Errorf("workspace (%s) is configured with invalid connection: %w", sourceWorkspace.Id(), err)
	}
	found := sourceWorkspace.FindRelativeConnection(*conn.EffectiveTarget)
	log.Printf("(getConnectionWorkspace) Found workspace in connections @ %s", found.Id())
	return &found, nil
}

func (d *dataConnection) getConnectionsFromRunConfig(runConfig *types.RunConfig) types.Connections {
	if runConfig == nil {
		return types.Connections{}
	}

	// If this is an app connection, we immediately return those
	if d.isAppConnection {
		return runConfig.Connections
	}
	// If the provider is configured with a non-empty capability name
	//   we should use the connections from that capability
	if capabilityName := d.p.PlanConfig.CapabilityName; capabilityName != "" {
		for _, cur := range runConfig.Capabilities {
			if cur.Name == capabilityName {
				return cur.Connections
			}
		}
		return types.Connections{}
	}
	return runConfig.Connections
}

func (d *dataConnection) validateConnection(conn types.Connection, wantContractName types.ModuleContractName, wantType string) error {
	// We are migrating from type = "..." to contract="..."
	// During migration, if the connection does not have a contract, then we won't perform any validation
	// Otherwise, if terraform has a contract, match against the connection
	// Else, match terraform type against the connection
	isEmptyContract := conn.Contract == "*/*/*" || conn.Contract == ""
	isEmptyType := conn.Type == "unknown" || conn.Type == ""
	if isEmptyContract && isEmptyType {
		return fmt.Errorf("there is no contract in the connection")
	} else if !isEmptyContract {
		if conn.Contract == "" {
			// Ignore contract validation if connection doesn't contain correct contract yet
			return nil
		}

		contractName, err := types.ParseModuleContractName(conn.Contract)
		if err != nil {
			return fmt.Errorf("the connection contract is invalid (%s): %s", conn.Contract, err)
		}
		if !wantContractName.Match(contractName) {
			return fmt.Errorf("the connection contract does not match (terraform match=%s, configured in Nullstone=%s)", wantContractName, conn.Contract)
		}
	} else {
		if conn.Type != wantType {
			return fmt.Errorf("the connection type does not match (terraform type=%s, configured in Nullstone=%s)", wantType, conn.Type)
		}
	}
	return nil
}
