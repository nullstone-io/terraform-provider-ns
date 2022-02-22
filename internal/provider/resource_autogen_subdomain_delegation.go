package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type resourceAutogenSubdomainDelegation struct {
	p *provider
}

func newResourceAutogenSubdomainDelegation(p *provider) (*resourceAutogenSubdomainDelegation, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &resourceAutogenSubdomainDelegation{p: p}, nil
}

var (
	_ server.Resource        = (*resourceAutogenSubdomainDelegation)(nil)
	_ server.ResourceUpdater = (*resourceAutogenSubdomainDelegation)(nil)
)

func (r *resourceAutogenSubdomainDelegation) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Block: &tfprotov5.SchemaBlock{
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "subdomain_id",
					Type:            tftypes.Number,
					Description:     "The autogen subdomain belongs to this subdomain.",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "env_id",
					Type:            tftypes.Number,
					Description:     "The autogen subdomain belongs to this env.",
					Required:        true,
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

func (r *resourceAutogenSubdomainDelegation) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (r *resourceAutogenSubdomainDelegation) PlanCreate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	// There are never any creates, just set the values to proposed
	return map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"subdomain_id": proposed["subdomain_id"],
		"env_id":       proposed["env_id"],
		"nameservers":  proposed["nameservers"],
	}, nil, nil
}

func (r *resourceAutogenSubdomainDelegation) PlanUpdate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	return map[string]tftypes.Value{
		"id":           prior["id"],
		"subdomain_id": proposed["subdomain_id"],
		"env_id":       proposed["env_id"],
		"nameservers":  proposed["nameservers"],
	}, nil, nil
}

func (r *resourceAutogenSubdomainDelegation) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomainId := extractInt64FromConfig(config, "subdomain_id")
	envId := extractInt64FromConfig(config, "env_id")

	nsClient := &api.Client{Config: r.p.NsConfig}
	autogenSubdomain, err := nsClient.AutogenSubdomain().Get(subdomainId, envId)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error retrieving autogen subdomain delegation",
			Detail:   err.Error(),
		})
	} else if autogenSubdomain == nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "unable to find autogen subdomain for the given subdomain and environment",
			Detail:   fmt.Sprintf("subdomain_id=%d env_id=%d", subdomainId, envId),
		})
	} else {
		state["id"] = tftypes.NewValue(tftypes.String, fmt.Sprintf("%d", autogenSubdomain.Id))
		state["subdomain_id"] = tftypes.NewValue(tftypes.Number, &subdomainId)
		state["env_id"] = tftypes.NewValue(tftypes.Number, &envId)
		state["nameservers"] = ns.NameserversToProtov5(autogenSubdomain.Nameservers)
	}

	return state, diags, nil
}

func (r *resourceAutogenSubdomainDelegation) Create(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error) {
	return r.Update(ctx, planned, config, prior)
}

func (r *resourceAutogenSubdomainDelegation) Update(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomainId := extractInt64FromConfig(config, "subdomain_id")
	envId := extractInt64FromConfig(config, "env_id")

	nameservers, _ := extractStringSliceFromConfig(planned, "nameservers")
	autogenSubdomain := &types.AutogenSubdomain{Nameservers: nameservers}

	var id int64

	nsClient := &api.Client{Config: r.p.NsConfig}
	if result, err := nsClient.AutogenSubdomainDelegation().Update(subdomainId, envId, autogenSubdomain); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error updating autogen subdomain delegation",
			Detail:   err.Error(),
		})
	} else if result == nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The autogen_subdomain_delegation for the subdomain %d and env %d is missing.", subdomainId, envId),
		})
	} else {
		id = autogenSubdomain.Id
		nameservers = autogenSubdomain.Nameservers
	}

	state["id"] = tftypes.NewValue(tftypes.String, fmt.Sprintf("%d", id))
	state["subdomain_id"] = tftypes.NewValue(tftypes.Number, &subdomainId)
	state["env_id"] = tftypes.NewValue(tftypes.Number, &envId)
	state["nameservers"] = ns.NameserversToProtov5(nameservers)

	return state, diags, nil
}

func (r *resourceAutogenSubdomainDelegation) Destroy(ctx context.Context, prior map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomainId := extractInt64FromConfig(prior, "subdomain_id")
	envId := extractInt64FromConfig(prior, "env_id")

	nsClient := &api.Client{Config: r.p.NsConfig}
	if found, err := nsClient.AutogenSubdomainDelegation().Destroy(subdomainId, envId); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error destroying autogen subdomain delegation",
			Detail:   err.Error(),
		})
	} else if !found {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The autogen_subdomain_delegation for the subdomain %d and env %d is missing.", subdomainId, envId),
		})
	}

	return diags, nil
}
