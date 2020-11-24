package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type DataSource interface {
	Schema(ctx context.Context) *tfprotov5.Schema
	Validate(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error)
	Read(ctx context.Context, config map[string]tftypes.Value) (state map[string]tftypes.Value, diags []*tfprotov5.Diagnostic, err error)
}
