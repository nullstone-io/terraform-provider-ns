package provider

import "fmt"

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
