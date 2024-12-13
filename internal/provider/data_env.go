package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

type dataEnv struct {
	p *provider
}

func newDataEnv(p *provider) (*dataEnv, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataEnv{p: p}, nil
}

func (*dataEnv) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to read the nullstone environment",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "stack_id",
					Type:            tftypes.Number,
					Description:     "The stack ID that owns this environment",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "env_id",
					Type:            tftypes.Number,
					Description:     "The environment ID",
					Required:        true,
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "name",
					Type:            tftypes.String,
					Description:     "The name of environment.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Computed:        true,
				},
				{
					Name:            "type",
					Type:            tftypes.String,
					Description:     "The type of environment. Possible values: PipelineEnv, PreviewEnv, PreviewsSharedEnv, GlobalEnv",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Computed:        true,
				},
				{
					Name:            "pipeline_order",
					Type:            tftypes.Number,
					Description:     "If a PipelineEnv, this is a number representing which order in the pipeline.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Computed:        true,
				},
			},
		},
	}
}

func (d *dataEnv) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataEnv) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	nsConfig := d.p.NsConfig
	nsClient := api.Client{Config: nsConfig}

	diags := make([]*tfprotov5.Diagnostic, 0)

	stackId := extractInt64FromConfig(config, "stack_id")
	envId := extractInt64FromConfig(config, "env_id")

	var envName string
	var envType string
	var pipelineOrder int

	env, err := nsClient.Environments().Get(ctx, stackId, envId, false)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unable to find nullstone environment.",
			Detail:   err.Error(),
		})
	} else if env != nil {
		envName = env.Name
		envType = string(env.Type)
		if env.PipelineOrder != nil {
			pipelineOrder = *env.PipelineOrder
		}
	} else {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The environment %d in the stack %d does not exist in nullstone.", stackId, envId),
		})
	}

	return map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, fmt.Sprintf("%d", envId)),
		"stack_id":       tftypes.NewValue(tftypes.Number, &stackId),
		"env_id":         tftypes.NewValue(tftypes.Number, &envId),
		"name":           tftypes.NewValue(tftypes.String, envName),
		"type":           tftypes.NewValue(tftypes.String, envType),
		"pipeline_order": tftypes.NewValue(tftypes.Number, pipelineOrder),
	}, diags, nil
}
