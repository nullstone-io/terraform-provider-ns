package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"stack": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NULLSTONE_STACK", ""),
				Description: "The name of the stack in nullstone that owns this workspace.",
			},
			"env": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NULLSTONE_ENV", ""),
				Description: "The name of the environment in nullstone associated with this workspace.",
			},
			"block": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NULLSTONE_BLOCK", ""),
				Description: "The name of the block in nullstone associated with this workspace.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "A default list of tags including all nullstone configuration for this workspace.",
			},
			"hyphenated_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A standard, unique, computed name for the workspace using '-' as a delimiter that is typically used for resource names.",
			},
			"slashed_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A standard, unique, computed name for the workspace using '/' as a delimiter that is typically used for resource names.",
			},
		},
	}
}

func dataSourceNsWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	stackName := d.Get("stack").(string)
	envName := d.Get("env").(string)
	blockName := d.Get("block").(string)
	d.Set("tags", map[string]string{
		"Stack": stackName,
		"Env":   envName,
		"Block": blockName,
	})
	hyphenatedName := fmt.Sprintf("%s-%s-%s", stackName, envName, blockName)
	d.Set("hyphenated_name", hyphenatedName)
	d.Set("slashed_name", fmt.Sprintf("%s/%s/%s", stackName, envName, blockName))
	d.SetId(hyphenatedName)
	return nil
}
