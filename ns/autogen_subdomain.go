package ns

import "github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"

type AutogenSubdomain struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	DomainName string `json:"domainName"`
}

type AutogenSubdomainDelegation struct {
	Nameservers Nameservers `json:"nameservers"`
}

type Nameservers []string

func (s Nameservers) ToProtov5() tftypes.Value {
	nameservers := make([]tftypes.Value, 0)
	for _, nameserver := range s {
		nameservers = append(nameservers, tftypes.NewValue(tftypes.String, nameserver))
	}
	return tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nameservers)
}
