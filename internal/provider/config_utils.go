package provider

import "github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"

func stringFromConfig(config map[string]tftypes.Value, key string) string {
	if config[key].IsNull() {
		return ""
	}
	val := ""
	config[key].As(&val)
	return val
}

func boolFromConfig(config map[string]tftypes.Value, key string) bool {
	if config[key].IsNull() {
		return false
	}
	val := false
	config[key].As(&val)
	return val
}

func stringSliceFromConfig(config map[string]tftypes.Value, key string) ([]string, error) {
	if config[key].IsNull() {
		return make([]string, 0), nil
	}

	tfslice := make([]tftypes.Value, 0)
	if err := config[key].As(&tfslice); err != nil {
		return nil, err
	}

	slice := make([]string, 0)
	for _, tfitem := range tfslice {
		var item string
		if err := tfitem.As(&item); err != nil {
			return nil, err
		}
		slice = append(slice, item)
	}
	return slice, nil
}
