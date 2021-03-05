package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
)

type dataAutogenSubdomain struct {
	p *provider
}

func newDataAutogenSubdomain(p *provider) (*dataAutogenSubdomain, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataAutogenSubdomain{p: p}, nil
}

var (
	_ server.DataSource = (*dataAutogenSubdomain)(nil)
)

func (*dataAutogenSubdomain) Schema(ctx context.Context) *tfprotov5.Schema {
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
					Description:     "The name of the autogenerated subdomain.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "domain_name",
					Type:            tftypes.String,
					Computed:        true,
					Description:     "The domain name that nullstone manages for this autogenerated subdomain. It is usually `nullstone.app`.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "fqdn",
					Type:            tftypes.String,
					Computed:        true,
					Description:     "The fully-qualified domain name (FQDN) that nullstone manages for this autogenerated subdomain. It is composed as `{name}.{domain_name}`.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (d *dataAutogenSubdomain) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := make([]*tfprotov5.Diagnostic, 0)

	var name string
	if err := config["name"].As(&name); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "name must be a string",
		})
	}

	return diags, nil
}

func (d *dataAutogenSubdomain) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	name := stringFromConfig(config, "name")

	state := map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, name),
		"name": tftypes.NewValue(tftypes.String, name),
	}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomain, err := d.p.NsClient.GetAutogenSubdomain(name)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error retrieving autogen subdomain",
			Detail:   err.Error(),
		})
	} else if subdomain == nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The autogen_subdomain %q is missing.", name),
		})
	} else {
		state["domain_name"] = tftypes.NewValue(tftypes.String, subdomain.DomainName)
		state["fqdn"] = tftypes.NewValue(tftypes.String, fmt.Sprintf("%s.%s", subdomain.Name, subdomain.DomainName))
	}

	return state, diags, nil
}
