package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

type dataAgent struct {
	p *provider
}

func newDataAgent(p *provider) (*dataAgent, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataAgent{p: p}, nil
}

func (*dataAgent) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to read info about the Nullstone Agent",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "aws_account_id",
					Type:            tftypes.String,
					Description:     "The AWS Account ID of the Nullstone Agent",
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "aws_role_name",
					Type:            tftypes.String,
					Description:     "The AWS Role Name of the Nullstone Agent",
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "aws_role_arn",
					Type:            tftypes.String,
					Description:     "The AWS Role ARN of the Nullstone Agent",
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "gcp_project_id",
					Type:            tftypes.String,
					Description:     "The GCP Project ID of the Nullstone Agent",
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "gcp_service_account_email",
					Type:            tftypes.String,
					Description:     "The GCP Service Account Email of the Nullstone Agent",
					Computed:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (d *dataAgent) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataAgent) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	nsConfig := d.p.NsConfig
	nsClient := api.Client{Config: nsConfig}

	diags := make([]*tfprotov5.Diagnostic, 0)

	var awsAccountId, awsRoleName, awsRoleArn string
	var gcpProjectId, gcpServiceAccountEmail string
	agentInfo, err := nsClient.NullstoneAgent().Get(ctx)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to find nullstone agent info.",
			Detail:   err.Error(),
		})
	} else if agentInfo != nil {
		awsAccountId = agentInfo.Aws.AccountId
		awsRoleName = agentInfo.Aws.RoleName
		awsRoleArn = agentInfo.Aws.RoleArn
		gcpServiceAccountEmail = agentInfo.Gcp.ServiceAccountEmail
		gcpProjectId = agentInfo.Gcp.ProjectId
	} else {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The API didn't read info on the Nullstone agent."),
		})
	}

	return map[string]tftypes.Value{
		"id":                        tftypes.NewValue(tftypes.String, "nullstone-agent"),
		"aws_account_id":            tftypes.NewValue(tftypes.String, &awsAccountId),
		"aws_role_name":             tftypes.NewValue(tftypes.String, &awsRoleName),
		"aws_role_arn":              tftypes.NewValue(tftypes.String, &awsRoleArn),
		"gcp_project_id":            tftypes.NewValue(tftypes.String, &gcpProjectId),
		"gcp_service_account_email": tftypes.NewValue(tftypes.String, &gcpServiceAccountEmail),
	}, diags, nil
}
