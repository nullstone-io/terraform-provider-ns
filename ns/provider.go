package ns

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{},
		DataSourcesMap: map[string]*schema.Resource{
			"ns_workspace": dataSourceNsWorkspace(),
		},
		ResourcesMap: map[string]*schema.Resource{},
	}
	return p
}
