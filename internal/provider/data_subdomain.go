package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

type dataSubdomain struct {
	p *provider
}

func newDataSubdomain(p *provider) (*dataSubdomain, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataSubdomain{p: p}, nil
}

func (*dataSubdomain) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to read a nullstone subdomain.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "stack_id",
					Type:            tftypes.Number,
					Description:     "The stack ID that owns this subdomain",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "block_id",
					Type:            tftypes.Number,
					Description:     "The block ID of the subdomain (in the specified stack)",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "env_id",
					Type:            tftypes.Number,
					Description:     "The env ID of the subdomain (in the specified stack)",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "dns_name",
					Type:            tftypes.String,
					Description:     `The DNS Name identified on the Subdomain block. FQDN = "<dns_name>[.<env-chunk>].<domain>."`,
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "subdomain_name",
					Type:            tftypes.String,
					Description:     `The subdomain identified on this subdomain workspace. FQDN = "<subdomain_name>.<domain>.". This is equivalent to "<dns_name>[.<env-chunk>]".`,
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "domain_name",
					Type:            tftypes.String,
					Description:     `The domain identified on the parent domain for this subdomain workspace. FQDN = "<subdomain_name>.<domain_name>".`,
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "fqdn",
					Type:            tftypes.String,
					Description:     "The FQDN identified on the Subdomain in the given workspace. NOTE: This has a trailing '.'.",
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (d *dataSubdomain) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataSubdomain) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	nsConfig := d.p.NsConfig
	nsClient := api.Client{Config: nsConfig}

	diags := make([]*tfprotov5.Diagnostic, 0)

	stackId := extractInt64FromConfig(config, "stack_id")
	blockId := extractInt64FromConfig(config, "block_id")
	envId := extractInt64FromConfig(config, "env_id")

	result := map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, ""),
		"stack_id":       config["stack_id"],
		"block_id":       config["block_id"],
		"env_id":         config["env_id"],
		"dns_name":       tftypes.NewValue(tftypes.String, ""),
		"subdomain_name": tftypes.NewValue(tftypes.String, ""),
		"domain_name":    tftypes.NewValue(tftypes.String, ""),
		"fqdn":           tftypes.NewValue(tftypes.String, ""),
	}

	subdomainWorkspace, err := nsClient.SubdomainWorkspaces().Get(ctx, stackId, blockId, envId)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to find nullstone subdomain workspace.",
			Detail:   err.Error(),
		})
	} else if subdomainWorkspace == nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The subdomain in the stack %d and block %d does not exist in nullstone.", stackId, blockId),
		})
	} else {
		result["id"] = tftypes.NewValue(tftypes.String, subdomainWorkspace.WorkspaceUid.String())
		result["dns_name"] = tftypes.NewValue(tftypes.String, subdomainWorkspace.DnsName)
		result["subdomain_name"] = tftypes.NewValue(tftypes.String, subdomainWorkspace.SubdomainName)
		result["domain_name"] = tftypes.NewValue(tftypes.String, subdomainWorkspace.DomainName)
		result["fqdn"] = tftypes.NewValue(tftypes.String, subdomainWorkspace.Fqdn)
	}

	return result, diags, nil
}
