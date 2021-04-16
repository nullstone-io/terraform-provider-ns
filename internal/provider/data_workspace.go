package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/ns"
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
					Name:            "workspace_id",
					Computed:        true,
					Description:     "The fully qualified workspace ID. This follows the form `<stack>/<env>/<block>`.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Type:            tftypes.String,
				},
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
	return nil, nil
}

func (d *dataWorkspace) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	envCurWorkspace := d.p.PlanConfig.WorkspaceLocation
	stack := extractStringFromConfig(config, "stack")
	if stack == "" {
		stack = envCurWorkspace.Stack
	}
	env := extractStringFromConfig(config, "env")
	if env == "" {
		env = envCurWorkspace.Env
	}
	block := extractStringFromConfig(config, "block")
	if block == "" {
		block = envCurWorkspace.Block
	}
	destWorkspace := &ns.WorkspaceLocation{
		Stack: stack,
		Env:   env,
		Block: block,
	}

	tags := map[string]tftypes.Value{
		"Stack": tftypes.NewValue(tftypes.String, stack),
		"Env":   tftypes.NewValue(tftypes.String, env),
		"Block": tftypes.NewValue(tftypes.String, block),
	}
	hyphenated := fmt.Sprintf("%s-%s-%s", stack, env, block)
	slashed := fmt.Sprintf("%s/%s/%s", stack, env, block)

	return map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, slashed),
		"workspace_id":    tftypes.NewValue(tftypes.String, destWorkspace.Id()),
		"stack":           tftypes.NewValue(tftypes.String, stack),
		"env":             tftypes.NewValue(tftypes.String, env),
		"block":           tftypes.NewValue(tftypes.String, block),
		"tags":            tftypes.NewValue(tftypes.Map{AttributeType: tftypes.String}, tags),
		"hyphenated_name": tftypes.NewValue(tftypes.String, hyphenated),
		"slashed_name":    tftypes.NewValue(tftypes.String, slashed),
	}, nil, nil
}
