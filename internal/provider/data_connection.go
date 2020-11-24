package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
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
					Name:            "workspace",
					Type:            tftypes.String,
					Computed:        true,
					Description:     "The name of the connected workspace.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:     "via",
					Type:     tftypes.String,
					Optional: true,
					Description: `Defines this connection is satisfied through another ns_connection.
Typically, this is set to data.ns_connection.other.workspace.`,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (d *dataConnection) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := make([]*tfprotov5.Diagnostic, 0)

	var name string
	if err := config["name"].As(&name); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "ns_connection.name must be a string",
		})
	} else if !validConnectionName.Match([]byte(name)) {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "ns_connection.name can only contain the characters 'a'-'z', '0'-'9', '-', '_'",
		})
	}

	var optional bool
	if err := config["optional"].As(&optional); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  err.Error(),
		})
	}

	workspace := os.Getenv(fmt.Sprintf("NULLSTONE_CONNECTION_%s", name))
	if workspace == "" && !optional {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The connection %q is missing. It is required to use this plan.", name),
		})
	}

	if len(diags) > 0 {
		return diags, nil
	}

	return nil, nil
}

func (d *dataConnection) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	var (
		name     string
		optional bool
	)

	if err := config["name"].As(&name); err != nil {
		return nil, nil, err
	}
	if err := config["optional"].As(&optional); err != nil {
		return nil, nil, err
	}

	workspace := os.Getenv(fmt.Sprintf("NULLSTONE_CONNECTION_%s", name))

	return map[string]tftypes.Value{
		"id":        tftypes.NewValue(tftypes.String, fmt.Sprintf("%s-%s", name, workspace)),
		"workspace": tftypes.NewValue(tftypes.String, workspace),
	}, nil, nil
}
