package provider

import (
	"fmt"
	"github.com/nullstone-io/terraform-provider-ns/internal/server"
)

var _ server.DataSource = &dataAppConnection{}

type dataAppConnection struct {
	dataConnection
}

func newDataAppConnection(p *provider) (*dataAppConnection, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataAppConnection{
		dataConnection: dataConnection{
			p:               p,
			isAppConnection: true,
		},
	}, nil
}
