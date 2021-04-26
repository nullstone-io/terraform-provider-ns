package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

type resourceAutogenSubdomain struct {
	p *provider
}

func newResourceAutogenSubdomain(p *provider) (*resourceAutogenSubdomain, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &resourceAutogenSubdomain{p: p}, nil
}

var (
	_ server.Resource        = (*resourceAutogenSubdomain)(nil)
	_ server.ResourceUpdater = (*resourceAutogenSubdomain)(nil)
)

func (r *resourceAutogenSubdomain) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Block: &tfprotov5.SchemaBlock{
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "subdomainId",
					Type:            tftypes.Number,
					Description:     "The autogen subdomain belongs to this subdomain.",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "env",
					Type:            tftypes.String,
					Description:     "The autogen subdomain belongs to this env.",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "dns_name",
					Type:            tftypes.String,
					Computed:        true,
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
					Description:     "The fully-qualified domain name (FQDN) that nullstone manages for this autogenerated subdomain. It is composed as `{name}.{domain_name}.`.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (r *resourceAutogenSubdomain) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (r *resourceAutogenSubdomain) PlanCreate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	// All values are computed, we set to UnknownValue to tell TF that we will change
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"dns_name":    tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"domain_name": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"fqdn":        tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	}, nil, nil
}

func (r *resourceAutogenSubdomain) PlanUpdate(ctx context.Context, proposed map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	// There are never any updates, just set the values to prior
	return map[string]tftypes.Value{
		"id":          prior["id"],
		"dns_name":    prior["dns_name"],
		"domain_name": prior["domain_name"],
		"fqdn":        prior["fqdn"],
	}, nil, nil
}

func (r *resourceAutogenSubdomain) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomainId := extractIntFromConfig(config, "subdomainId")
	envName := extractStringFromConfig(config,"env")

	nsClient := &api.Client{Config: r.p.NsConfig}
	autogenSubdomain, err := nsClient.AutogenSubdomain().Get(subdomainId, envName)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error retrieving autogen subdomain",
			Detail:   err.Error(),
		})
	} else if autogenSubdomain == nil {
		state["id"] = tftypes.NewValue(tftypes.Number, 0)
		state["dns_name"] = tftypes.NewValue(tftypes.String, "")
		state["domain_name"] = tftypes.NewValue(tftypes.String, "")
		state["fqdn"] = tftypes.NewValue(tftypes.String, "")
	} else {
		state["id"] = tftypes.NewValue(tftypes.Number, autogenSubdomain.Id)
		state["dns_name"] = tftypes.NewValue(tftypes.String, autogenSubdomain.DnsName)
		state["domain_name"] = tftypes.NewValue(tftypes.String, autogenSubdomain.DomainName)
		state["fqdn"] = tftypes.NewValue(tftypes.String, autogenSubdomain.Fqdn)
	}

	return state, diags, nil
}

func (r *resourceAutogenSubdomain) Create(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomainId := extractIntFromConfig(config, "subdomainId")
	envName := extractStringFromConfig(config,"env")

	nsClient := &api.Client{Config: r.p.NsConfig}
	if autogenSubdomain, err := nsClient.AutogenSubdomain().Create(subdomainId, envName); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error creating autogen subdomain",
			Detail:   err.Error(),
		})
	} else {
		state["id"] = tftypes.NewValue(tftypes.Number, autogenSubdomain.Id)
		state["dns_name"] = tftypes.NewValue(tftypes.String, autogenSubdomain.DnsName)
		state["domain_name"] = tftypes.NewValue(tftypes.String, autogenSubdomain.DomainName)
		state["fqdn"] = tftypes.NewValue(tftypes.String, autogenSubdomain.Fqdn)
	}

	return state, diags, nil
}

func (r *resourceAutogenSubdomain) Update(ctx context.Context, planned map[string]tftypes.Value, config map[string]tftypes.Value, prior map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	// NOTE: AutogenSubdomains cannot be updated, this is going to do nothing
	state := map[string]tftypes.Value{}
	diags := make([]*tfprotov5.Diagnostic, 0)

	state["id"] = config["id"]
	state["dns_name"] = config["dns_name"]
	state["domain_name"] = config["domain_name"]
	state["fqdn"] = config["fqdn"]

	return state, diags, nil
}

func (r *resourceAutogenSubdomain) Destroy(ctx context.Context, prior map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := make([]*tfprotov5.Diagnostic, 0)

	subdomainId := extractIntFromConfig(config, "subdomainId")
	envName := extractStringFromConfig(config,"env")

	nsClient := &api.Client{Config: r.p.NsConfig}
	if found, err := nsClient.AutogenSubdomain().Destroy(subdomainId, envName); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "error destroying autogen subdomain",
			Detail:   err.Error(),
		})
	} else if !found {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The autogen_subdomain %q is missing.", name),
		})
	}

	return diags, nil
}
