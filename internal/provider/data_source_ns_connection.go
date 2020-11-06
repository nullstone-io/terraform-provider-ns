package provider

import (
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var validConnectionName = regexp.MustCompile("^[_a-z0-9/-]+$")

func dataSourceNsConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsConnectionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique name of the connection within this module.",
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					val, ok := i.(string)
					if !ok {
						return diag.Errorf("ns_connection.name must be a string")
					}
					if !validConnectionName.Match([]byte(val)) {
						return diag.Errorf("ns_connection.name can only contain the characters 'a'-'z', '0'-'9', '-', '_'")
					}
					return nil
				},
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of module to satisfy this connection.",
			},
			"optional": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "This data source will cause an error if optional is false and this connection is not configured.",
			},
			"workspace": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the connected workspace.",
			},
		},
	}
}

func dataSourceNsConnectionRead(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	optional := d.Get("optional").(bool)

	workspace := os.Getenv(fmt.Sprintf("NULLSTONE_CONNECTION_%s", name))
	if workspace == "" && !optional {
		return fmt.Errorf("The connection %q is missing. It is required to use this plan.", name)
	}
	d.Set("workspace", workspace)
	d.SetId(fmt.Sprintf("%s-%s", name, workspace))

	return nil
}
