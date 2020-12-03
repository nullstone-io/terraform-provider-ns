package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type dataWorkspace struct {
	p *provider
}

func newDataWorkspace(p *provider) (*dataWorkspace, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataWorkspace{p: p}, nil
}

func (*dataWorkspace) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to configure module based on current nullstone workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "stack",
					Type:            tftypes.String,
					Description:     "The name of the stack in nullstone that owns this workspace.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "env",
					Type:            tftypes.String,
					Description:     "The name of the environment in nullstone associated with this workspace.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "block",
					Type:            tftypes.String,
					Description:     "The name of the block in nullstone associated with this workspace.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "tags",
					Type:            tftypes.Map{AttributeType: tftypes.String},
					Computed:        true,
					Description:     "A default list of tags including all nullstone configuration for this workspace.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "hyphenated_name",
					Type:            tftypes.String,
					Computed:        true,
					Description:     "A standard, unique, computed name for the workspace using '-' as a delimiter that is typically used for resource names.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "slashed_name",
					Type:            tftypes.String,
					Computed:        true,
					Description:     "A standard, unique, computed name for the workspace using '/' as a delimiter that is typically used for resource names.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (d *dataWorkspace) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	diags := make([]*tfprotov5.Diagnostic, 0)

	if !config["stack"].IsNull() {
		var stack string
		if err := config["stack"].As(&stack); err != nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  err.Error(),
			})
		}
	}
	if !config["env"].IsNull() {
		var env string
		if err := config["env"].As(&env); err != nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  err.Error(),
			})
		}
	}
	if !config["block"].IsNull() {
		var block string
		if err := config["block"].As(&block); err != nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  err.Error(),
			})
		}
	}

	if len(diags) > 0 {
		return diags, nil
	}
	return nil, nil
}

func (d *dataWorkspace) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	stack := stringFromConfig(config, "stack")
	if stack == "" {
		stack = d.p.PlanConfig.Stack
	}
	env := stringFromConfig(config, "env")
	if env == "" {
		env = d.p.PlanConfig.Env
	}
	block := stringFromConfig(config, "block")
	if block == "" {
		block = d.p.PlanConfig.Block
	}

	tags := map[string]tftypes.Value{
		"Stack": tftypes.NewValue(tftypes.String, stack),
		"Env":   tftypes.NewValue(tftypes.String, env),
		"Block": tftypes.NewValue(tftypes.String, block),
	}
	hyphenated := fmt.Sprintf("%s-%s-%s", stack, env, block)
	slashed := fmt.Sprintf("%s/%s/%s", stack, env, block)

	return map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, hyphenated),
		"stack":           tftypes.NewValue(tftypes.String, stack),
		"env":             tftypes.NewValue(tftypes.String, env),
		"block":           tftypes.NewValue(tftypes.String, block),
		"tags":            tftypes.NewValue(tftypes.Map{AttributeType: tftypes.String}, tags),
		"hyphenated_name": tftypes.NewValue(tftypes.String, hyphenated),
		"slashed_name":    tftypes.NewValue(tftypes.String, slashed),
	}, nil, nil
}
