package provider

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"math/big"
	"regexp"
)

func extractStringFromTfValue(tfvalue tftypes.Value) string {
	if tfvalue.IsNull() {
		return ""
	}
	val := ""
	tfvalue.As(&val)
	return val
}

func extractStringFromConfig(config map[string]tftypes.Value, key string) string {
	if config[key].IsNull() {
		return ""
	}
	val := ""
	config[key].As(&val)
	return val
}

func extractBoolFromConfig(config map[string]tftypes.Value, key string) bool {
	if config[key].IsNull() {
		return false
	}
	val := false
	config[key].As(&val)
	return val
}

func extractIntFromConfig(config map[string]tftypes.Value, key string) int {
	if config[key].IsNull() {
		return -1
	}
	val := new(big.Float)
	config[key].As(&val)
	i, _ := val.Int64()
	return int(i)
}

func extractInt64FromConfig(config map[string]tftypes.Value, key string) int64 {
	if config[key].IsNull() {
		return -1
	}
	val := new(big.Float)
	config[key].As(&val)
	i, _ := val.Int64()
	return i
}

func extractStringSliceFromConfig(config map[string]tftypes.Value, key string) ([]string, error) {
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

func extractMapFromConfig(config map[string]tftypes.Value, key string) map[string]tftypes.Value {
	if config[key].IsNull() {
		return make(map[string]tftypes.Value)
	}
	val := make(map[string]tftypes.Value)
	if err := config[key].As(&val); err != nil {
		return make(map[string]tftypes.Value)
	}
	return val
}

func copyMap(m map[string]tftypes.Value) map[string]tftypes.Value {
	copy := make(map[string]tftypes.Value)
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

const envVariableKeyRegex = "^[a-zA-Z_][a-zA-Z0-9_]*$"

func validEnvVariableKey(key string) bool {
	regex := regexp.MustCompile(envVariableKeyRegex)
	return regex.MatchString(key)
}
