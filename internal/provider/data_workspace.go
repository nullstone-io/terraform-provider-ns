package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
	attrs := []*tfprotov5.SchemaAttribute{
		deprecatedIDAttribute(),
		{
			Name:            "stack_id",
			Type:            tftypes.Number,
			Description:     "The ID of the stack in nullstone that owns this workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "stack_name",
			Type:            tftypes.String,
			Description:     "The name of the stack in nullstone that owns this workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "block_id",
			Type:            tftypes.Number,
			Description:     "The ID of the block in nullstone associated with this workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "block_name",
			Type:            tftypes.String,
			Description:     "The name of the block in nullstone that owns this workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name: "block_ref",
			Type: tftypes.String,
			Description: `The reference of the block in nullstone that owns this workspace.
This is typically used to construct unique resource names. See unique_name.`,
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "env_id",
			Type:            tftypes.Number,
			Description:     "The ID of the environment in nullstone associated with this workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "env_name",
			Type:            tftypes.String,
			Description:     "The name of the block in nullstone that owns this workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Computed:        true,
		},
		{
			Name:            "tags",
			Type:            tftypes.Map{ElementType: tftypes.String},
			Computed:        true,
			Description:     "A default list of tags including all nullstone configuration for this workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
		},
	}

	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to configure module based on current nullstone workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes:      attrs,
		},
	}
}

func (d *dataWorkspace) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataWorkspace) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	planConfig := d.p.PlanConfig

	stackId := extractInt64FromConfig(config, "stack_id")
	if stackId <= 0 {
		stackId = planConfig.StackId
	}
	stackName := extractStringFromConfig(config, "stack_name")
	if stackName == "" {
		stackName = planConfig.StackName
	}

	blockId := extractInt64FromConfig(config, "block_id")
	if blockId <= 0 {
		blockId = planConfig.BlockId
	}
	blockName := extractStringFromConfig(config, "block_name")
	if blockName == "" {
		blockName = planConfig.BlockName
	}
	blockRef := extractStringFromConfig(config, "block_ref")
	if blockRef == "" {
		blockRef = planConfig.BlockRef
	}

	envId := extractInt64FromConfig(config, "env_id")
	if envId <= 0 {
		envId = planConfig.EnvId
	}
	envName := extractStringFromConfig(config, "env_name")
	if envName == "" {
		envName = planConfig.EnvName
	}

	id := fmt.Sprintf("%s/%s/%s", stackName, blockName, envName)
	tags := map[string]tftypes.Value{
		"Stack": tftypes.NewValue(tftypes.String, stackName),
		"Env":   tftypes.NewValue(tftypes.String, envName),
		"Block": tftypes.NewValue(tftypes.String, blockName),
	}
	hyphenated := fmt.Sprintf("%s-%s-%s", stackName, envName, blockName)
	slashed := fmt.Sprintf("%s/%s/%s", stackName, envName, blockName)

	return map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, id),
		"stack_id":   tftypes.NewValue(tftypes.Number, &stackId),
		"stack_name": tftypes.NewValue(tftypes.String, stackName),
		"block_id":   tftypes.NewValue(tftypes.Number, &blockId),
		"block_name": tftypes.NewValue(tftypes.String, blockName),
		"block_ref":  tftypes.NewValue(tftypes.String, blockRef),
		"env_id":     tftypes.NewValue(tftypes.Number, &envId),
		"env_name":   tftypes.NewValue(tftypes.String, envName),
		"tags":       tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tags),

		// Deprecated
		"workspace_id":    tftypes.NewValue(tftypes.String, id),
		"stack":           tftypes.NewValue(tftypes.String, stackName),
		"block":           tftypes.NewValue(tftypes.String, blockName),
		"env":             tftypes.NewValue(tftypes.String, envName),
		"hyphenated_name": tftypes.NewValue(tftypes.String, hyphenated),
		"slashed_name":    tftypes.NewValue(tftypes.String, slashed),
	}, nil, nil
}
