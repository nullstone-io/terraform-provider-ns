package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"strconv"
)

type dataDomain struct {
	p *provider
}

func newDataDomain(p *provider) (*dataDomain, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataDomain{p: p}, nil
}

func (*dataDomain) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to read a nullstone domain.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "stack",
					Type:            tftypes.String,
					Description:     "The domain belongs to this stack",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "block",
					Type:            tftypes.String,
					Description:     "The domain belongs to this block (in the specified stack)",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "dns_name",
					Type:            tftypes.String,
					Description:     "The DNS name defined on the domain",
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (d *dataDomain) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataDomain) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	nsConfig := d.p.NsConfig
	nsClient := api.Client{Config: nsConfig}

	diags := make([]*tfprotov5.Diagnostic, 0)

	stack := extractStringFromConfig(config, "stack")
	block := extractStringFromConfig(config, "block")

	var domainId int
	var dnsName string

	domain, err := nsClient.Domains().Get(stack, block)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to find nullstone domain.",
			Detail:   err.Error(),
		})
	} else if domain != nil {
		domainId = domain.Id
		dnsName = domain.DnsName
	} else {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The domain in the stack %q and block %q does not exist in nullstone.", stack, block),
		})
	}

	return map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, strconv.Itoa(domainId)),
		"stack":    tftypes.NewValue(tftypes.String, stack),
		"block":    tftypes.NewValue(tftypes.String, block),
		"dns_name": tftypes.NewValue(tftypes.String, dnsName),
	}, diags, nil
}
