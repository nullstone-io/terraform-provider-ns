package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"strconv"
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
					Name:            "stack",
					Type:            tftypes.String,
					Description:     "The subdomain belongs to this stack",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "block",
					Type:            tftypes.String,
					Description:     "The subdomain belongs to this block (in the specified stack)",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "dns_name",
					Type:            tftypes.String,
					Description:     "The DNS name defined on the subdomain",
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

	stack := extractStringFromConfig(config, "stack")
	block := extractStringFromConfig(config, "block")

	var subdomainId int
	var dnsName string

	subdomain, err := nsClient.Subdomains().Get(stack, block)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to find nullstone subdomain.",
			Detail:   err.Error(),
		})
	} else if subdomain != nil {
		subdomainId = subdomain.Id
		dnsName = subdomain.DnsName
	} else {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The subdomain in the stack %q and block %q does not exist in nullstone.", stack, block),
		})
	}

	return map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, strconv.Itoa(subdomainId)),
		"stack":    tftypes.NewValue(tftypes.String, stack),
		"block":    tftypes.NewValue(tftypes.String, block),
		"dns_name": tftypes.NewValue(tftypes.String, dnsName),
	}, diags, nil
}
