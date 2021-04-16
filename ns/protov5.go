package ns

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func NameserversToProtov5(s types.Nameservers) tftypes.Value {
	nameservers := make([]tftypes.Value, 0)
	for _, nameserver := range s {
		nameservers = append(nameservers, tftypes.NewValue(tftypes.String, nameserver))
	}
	return tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nameservers)
}
