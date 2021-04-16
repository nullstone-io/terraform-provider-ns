package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type resourceSubdomainDelegation struct {
	p *provider
}

func newResourceSubdomainDelegation(p *provider) (*resourceSubdomainDelegation, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &resourceSubdomainDelegation{p: p}, nil
}

var (
	_ server.Resource        = (*resourceSubdomainDelegation)(nil)
	_ server.ResourceUpdater = (*resourceSubdomainDelegation)(nil)
)

func (r *resourceSubdomainDelegation) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Block: &tfprotov5.SchemaBlock{
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "subdomain",
					Required:        true,
					Type:            tftypes.String,
					Description:     "Name of auto-generated subdomain that already exists in Nullstone system. This should not include `nullstone.app`.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "nameservers",
					Required:        true,
					Type:            tftypes.List{ElementType: tftypes.String},
					Description:     "A list of nameservers that refer to a DNS zone where this subdomain can delegate.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (r *resourceSubdomainDelegation) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (r *resourceSubdomainDelegation) PlanCreate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	return r.plan(ctx, proposed)
}

func (r *resourceSubdomainDelegation) PlanUpdate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	return r.plan(ctx, proposed)
}

func (r *resourceSubdomainDelegation) plan(ctx context.Context, proposed map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	subdomainName := extractStringFromConfig(proposed, "subdomain")

	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, subdomainName),
		"subdomain":   proposed["subdomain"],
		"nameservers": proposed["nameservers"],
	}, nil, nil
}

func (r *resourceSubdomainDelegation) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomainName := extractStringFromConfig(config, "subdomain")

	nsClient := &api.Client{Config: r.p.NsConfig}
	delegation, err := nsClient.AutogenSubdomainsDelegation().Get(subdomainName)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error retrieving autogen subdomain delegation",
			Detail:   err.Error(),
		})
	} else {
		state["id"] = config["subdomain"]
		state["subdomain"] = config["subdomain"]
		state["nameservers"] = ns.NameserversToProtov5(delegation.Nameservers)
	}

	return state, diags, nil
}

func (r *resourceSubdomainDelegation) Create(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error) {
	return r.Update(ctx, planned, config, prior)
}

func (r *resourceSubdomainDelegation) Update(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomain := extractStringFromConfig(planned, "subdomain")
	nameservers, _ := extractStringSliceFromConfig(planned, "nameservers")
	delegation := &types.AutogenSubdomainDelegation{Nameservers: types.Nameservers(nameservers)}

	nsClient := &api.Client{Config: r.p.NsConfig}
	if result, err := nsClient.AutogenSubdomainsDelegation().Update(subdomain, delegation); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error updating autogen subdomain delegation",
			Detail:   err.Error(),
		})
	} else if result == nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The autogen_subdomain_delegation %q is missing.", subdomain),
		})
	} else {
		state["id"] = tftypes.NewValue(tftypes.String, subdomain)
		state["subdomain"] = tftypes.NewValue(tftypes.String, subdomain)
		state["nameservers"] = ns.NameserversToProtov5(result.Nameservers)
	}

	return state, diags, nil
}

func (r *resourceSubdomainDelegation) Destroy(ctx context.Context, prior map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomain := extractStringFromConfig(prior, "subdomain")
	nsClient := &api.Client{Config: r.p.NsConfig}
	if found, err := nsClient.AutogenSubdomainsDelegation().Destroy(subdomain); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error destroying autogen subdomain delegation",
			Detail:   err.Error(),
		})
	} else if !found {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The autogen_subdomain_delegation %q is missing.", subdomain),
		})
	}

	return diags, nil
}
