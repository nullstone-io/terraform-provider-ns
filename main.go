package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/nullstone-io/terraform-provider-ns/ns"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ns.Provider})
}
