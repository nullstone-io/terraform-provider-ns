package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
	"github.com/nullstone-io/terraform-provider-ns/ns"
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
					Description:     "",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "nameservers",
					Required:        true,
					Type:            tftypes.List{ElementType: tftypes.String},
					Description:     "",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (r *resourceSubdomainDelegation) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := make([]*tfprotov5.Diagnostic, 0)

	var subdomain string
	if err := config["subdomain"].As(&subdomain); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "subdomain must be a string",
		})
	}
	if _, err := stringSliceFromConfig(config, "nameservers"); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "invalid nameservers, must be list(string)",
			Detail:   err.Error(),
		})
	}

	return diags, nil
}

func (r *resourceSubdomainDelegation) PlanCreate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	return r.plan(ctx, proposed)
}

func (r *resourceSubdomainDelegation) PlanUpdate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	return r.plan(ctx, proposed)
}

func (r *resourceSubdomainDelegation) plan(ctx context.Context, proposed map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	subdomainName := stringFromConfig(proposed, "subdomain")

	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, subdomainName),
		"subdomain":   proposed["subdomain"],
		"nameservers": proposed["nameservers"],
	}, nil, nil
}

func (r *resourceSubdomainDelegation) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomainName := stringFromConfig(config, "subdomain")

	delegation, err := r.p.NsClient.GetAutogenSubdomainDelegation(subdomainName)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error retrieving autogen subdomain delegation",
			Detail:   err.Error(),
		})
	} else {
		state["id"] = config["subdomain"]
		state["subdomain"] = config["subdomain"]
		state["nameservers"] = delegation.Nameservers.ToProtov5()
	}

	return state, diags, nil
}

func (r *resourceSubdomainDelegation) Destroy(ctx context.Context, prior map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	planned := map[string]tftypes.Value{}
	config := map[string]tftypes.Value{
		"subdomain":   prior["subdomain"],
		"nameservers": ns.Nameservers{}.ToProtov5(),
	}
	_, diags, err := r.Update(ctx, planned, config, prior)
	return diags, err
}

func (r *resourceSubdomainDelegation) Create(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error) {
	return r.Update(ctx, planned, config, prior)
}

func (r *resourceSubdomainDelegation) Update(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomain := stringFromConfig(planned, "subdomain")
	nameservers, _ := stringSliceFromConfig(planned, "nameservers")
	delegation := &ns.AutogenSubdomainDelegation{Nameservers: ns.Nameservers(nameservers)}

	if result, err := r.p.NsClient.UpdateAutogenSubdomainDelegation(subdomain, delegation); err != nil {
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
		state["nameservers"] = result.Nameservers.ToProtov5()
	}

	return state, diags, nil
}
