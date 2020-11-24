package provider

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

// TODO: remove this once its not needed by testing
func deprecatedIDAttribute() *tfprotov5.SchemaAttribute {
	return &tfprotov5.SchemaAttribute{
		Name:       "id",
		Computed:   true,
		Deprecated: true,
		Description: "This attribute is only present for some compatibility issues and should not be used. It " +
			"will be removed in a future version.",
		DescriptionKind: tfprotov5.StringKindMarkdown,
		Type:            tftypes.String,
	}
}
