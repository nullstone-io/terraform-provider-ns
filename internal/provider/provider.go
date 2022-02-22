package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
	"github.com/nullstone-io/terraform-provider-ns/ns"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"log"
)

func Mock(version string, getNsConfig func() api.Config, getTfeConfig func() *tfe.Config) tfprotov5.ProviderServer {
	return newProviderServer(version, func() (api.Config, *tfe.Config, PlanConfig) {
		apiConfig := getNsConfig()
		tfeConfig := getTfeConfig()
		planConfig, _ := LoadPlanConfig()
		return apiConfig, tfeConfig, planConfig
	})
}

func New(version string) tfprotov5.ProviderServer {
	return newProviderServer(version, func() (api.Config, *tfe.Config, PlanConfig) {
		apiConfig := api.DefaultConfig()
		if profile, ac, _ := ns.LoadProfile(); profile != nil {
			apiConfig = ac
		}
		tfeConfig := ns.NewTfeConfig(apiConfig)
		planConfig, _ := LoadPlanConfig()
		return apiConfig, tfeConfig, planConfig
	})
}

func newProviderServer(version string, fn func() (api.Config, *tfe.Config, PlanConfig)) tfprotov5.ProviderServer {
	s := server.MustNew(func() server.Provider {
		apiConfig, tfeConfig, planConfig := fn()
		return &provider{
			Version:    version,
			NsConfig:   apiConfig,
			TfeConfig:  tfeConfig,
			PlanConfig: &planConfig,
		}
	})

	// data sources
	s.MustRegisterDataSource("ns_workspace", newDataWorkspace)
	s.MustRegisterDataSource("ns_connection", newDataConnection)
	s.MustRegisterDataSource("ns_app_connection", newDataAppConnection)
	s.MustRegisterDataSource("ns_subdomain", newDataSubdomain)
	s.MustRegisterDataSource("ns_domain", newDataDomain)
	s.MustRegisterDataSource("ns_app_env", newDataAppEnv)

	// resources
	s.MustRegisterResource("ns_autogen_subdomain", newResourceAutogenSubdomain)
	s.MustRegisterResource("ns_autogen_subdomain_delegation", newResourceAutogenSubdomainDelegation)

	return s
}

var _ server.Provider = (*provider)(nil)

type provider struct {
	Version    string
	TfeConfig  *tfe.Config
	TfeClient  *tfe.Client
	NsConfig   api.Config
	PlanConfig *PlanConfig
}

func (p *provider) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Attributes: []*tfprotov5.SchemaAttribute{
				{
					Name:            "organization",
					Type:            tftypes.String,
					Optional:        true,
					Description:     "Configure provider with this organization.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "capability_id",
					Type:            tftypes.Number,
					Optional:        true,
					Description:     "Configure provider with the context of the capability's id",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (p *provider) Validate(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error) {
	if !config["organization"].IsNull() {
		var orgName string
		if err := config["organization"].As(&orgName); err != nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "organization must be a string",
			})
		}
	}
	if p.NsConfig.ApiKey == "" {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("Nullstone API Key is required (Set %q environment variable)", api.ApiKeyEnvVar),
		})
	}
	if p.TfeConfig.Token == "" {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("TFE Token is required (Set %q environment variable)", api.ApiKeyEnvVar),
		})
	}

	if len(diags) > 0 {
		return diags, nil
	}

	return nil, nil
}

func (p *provider) Configure(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error) {
	log.Printf("[DEBUG] Configuring Nullstone provider %s", p.Version)
	if !config["organization"].IsNull() {
		// This is already checked in Validate, just cast it
		config["organization"].As(&p.PlanConfig.OrgName)
	}

	p.NsConfig.OrgName = p.PlanConfig.OrgName
	log.Printf("[DEBUG] Configured Nullstone API client (Address=%s)\n", p.NsConfig.BaseAddress)

	p.PlanConfig.CapabilityId = extractInt64FromConfig(config, "capability_id")
	log.Printf("[DEBUG] capability_id set to %d\n", p.PlanConfig.CapabilityId)

	p.TfeClient, err = tfe.NewClient(p.TfeConfig)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Configured TFE client (Address=%s, BasePath=%s)\n", p.TfeConfig.Address, p.TfeConfig.BasePath)

	return nil, nil
}
