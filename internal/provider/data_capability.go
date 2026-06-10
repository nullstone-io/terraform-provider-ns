package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type dataCapability struct {
	p *provider
}

func newDataCapability(p *provider) (*dataCapability, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataCapability{p: p}, nil
}

func (*dataCapability) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to read the current nullstone capability.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				{
					Name:            "id",
					Type:            tftypes.Number,
					Description:     "The ID of the capability this module is deployed as.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Computed:        true,
				},
				{
					Name:            "name",
					Type:            tftypes.String,
					Description:     "The name of the capability this module is deployed as.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
					Computed:        true,
				},
			},
		},
	}
}

func (d *dataCapability) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataCapability) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	planConfig := d.p.PlanConfig

	capabilityId := planConfig.CapabilityId

	return map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.Number, &capabilityId),
		"name": tftypes.NewValue(tftypes.String, planConfig.CapabilityName),
	}, nil, nil
}
