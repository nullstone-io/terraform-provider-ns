package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
)

func New(version string) func() tfprotov5.ProviderServer {
	return func() tfprotov5.ProviderServer {
		s := server.MustNew(func() server.Provider {
			return &provider{}
		})

		// data sources
		s.MustRegisterDataSource("ns_workspace", newDataWorkspace)
		s.MustRegisterDataSource("ns_connection", newDataConnection)

		return s
	}
}

var _ server.Provider = (*provider)(nil)

type provider struct {

}

func (p provider) Schema(ctx context.Context) *tfprotov5.Schema {
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Attributes: []*tfprotov5.SchemaAttribute{},
		},
	}
}

func (p provider) Validate(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error) {
	return nil, nil
}

func (p provider) Configure(ctx context.Context, config map[string]tftypes.Value) (diags []*tfprotov5.Diagnostic, err error) {
	return nil, nil
}
