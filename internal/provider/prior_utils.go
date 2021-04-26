package provider

import "github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"

func extractStringFromPrior(prior map[string]tftypes.Value, key string) string {
	if prior[key].IsNull() {
		return ""
	}
	val := ""
	prior[key].As(&val)
	return val
}

func extractIntFromPrior(prior map[string]tftypes.Value, key string) int {
	if prior[key].IsNull() {
		return -1
	}
	val := 0
	prior[key].As(&val)
	return val
}
