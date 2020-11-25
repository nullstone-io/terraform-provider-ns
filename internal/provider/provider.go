package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
	"github.com/nullstone-io/terraform-provider-ns/ns"
)

func New(version string, getTfeConfig func() *tfe.Config) tfprotov5.ProviderServer {
	s := server.MustNew(func() server.Provider {
		if getTfeConfig == nil {
			getTfeConfig = ns.NewTfeConfig
		}
		return &provider{TfeConfig: getTfeConfig()}
	})

	// data sources
	s.MustRegisterDataSource("ns_workspace", newDataWorkspace)
	s.MustRegisterDataSource("ns_connection", newDataConnection)

	return s
}

var _ server.Provider = (*provider)(nil)

type provider struct {
	TfeConfig *tfe.Config
	TfeClient *tfe.Client
	OrgName   string
}

func (p *provider) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Attributes: []*tfprotov5.SchemaAttribute{
				{
					Name:            "organization",
					Type:            tftypes.String,
					Description:     "Configure provider with this organization.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (p *provider) Validate(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error) {
	if _, err := p.getOrg(config); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  err.Error(),
		})
	}

	if len(diags) > 0 {
		return diags, nil
	}

	return nil, nil
}

func (p *provider) Configure(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error) {
	if p.OrgName, err = p.getOrg(config); err != nil {
		return nil, err
	}

	p.TfeClient, err = tfe.NewClient(p.TfeConfig)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (p *provider) getOrg(config map[string]tftypes.Value) (string, error) {
	var orgName string
	if err := config["organization"].As(&orgName); err != nil {
		return "", fmt.Errorf("organization must be a string")
	} else if orgName == "" {
		if orgName = os.Getenv("NULLSTONE_ORG"); orgName == "" {
			return "", fmt.Errorf("ns provider organization is required")
		}
	}
	return orgName, nil
}
