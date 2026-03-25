package provider

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"math/big"
	"regexp"
)

func MapToTfValue(m map[string]string) tftypes.Value {
	tfMap := map[string]tftypes.Value{}
	for k, v := range m {
		tfMap[k] = tftypes.NewValue(tftypes.String, v)
	}
	return tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tfMap)
}

func TfValueToMap(tfVal tftypes.Value) map[string]string {
	result := map[string]string{}
	if tfVal.IsNull() {
		return result
	}

	temp := make(map[string]tftypes.Value)
	if err := tfVal.As(&temp); err != nil {
		return result
	}

	for k, tfv := range temp {
		result[k] = extractStringFromTfValue(tfv)
	}
	return result
}

func TfSetValueToStringSlice(tfVal tftypes.Value) []string {
	result := make([]string, 0)
	if tfVal.IsNull() {
		return result
	}
	temp := make([]tftypes.Value, 0)
	if err := tfVal.As(&temp); err != nil {
		return result
	}
	for _, tfv := range temp {
		result = append(result, extractStringFromTfValue(tfv))
	}
	return result
}

func SliceToTfSet(s []string) tftypes.Value {
	tfSlice := make([]tftypes.Value, 0)
	for _, v := range s {
		tfSlice = append(tfSlice, tftypes.NewValue(tftypes.String, v))
	}
	return tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, tfSlice)
}

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

var fieldRefObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"api_version": tftypes.String,
	"field_path":  tftypes.String,
}}

var configMapRefObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"key":      tftypes.String,
	"name":     tftypes.String,
	"optional": tftypes.Bool,
}}

var resourceFieldRefObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"resource":  tftypes.String,
	"container": tftypes.String,
	"divisor":   tftypes.String,
}}

var fileKeyRefObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"key":         tftypes.String,
	"path":        tftypes.String,
	"volume_name": tftypes.String,
}}

func FieldRefsToTfValue(refs map[string]FieldRef) tftypes.Value {
	tfMap := map[string]tftypes.Value{}
	for k, v := range refs {
		tfMap[k] = tftypes.NewValue(fieldRefObjectType, map[string]tftypes.Value{
			"api_version": tftypes.NewValue(tftypes.String, v.ApiVersion),
			"field_path":  tftypes.NewValue(tftypes.String, v.FieldPath),
		})
	}
	return tftypes.NewValue(tftypes.Map{ElementType: fieldRefObjectType}, tfMap)
}

func ConfigMapRefsToTfValue(refs map[string]ConfigMapRef) tftypes.Value {
	tfMap := map[string]tftypes.Value{}
	for k, v := range refs {
		tfMap[k] = tftypes.NewValue(configMapRefObjectType, map[string]tftypes.Value{
			"key":      tftypes.NewValue(tftypes.String, v.Key),
			"name":     tftypes.NewValue(tftypes.String, v.Name),
			"optional": tftypes.NewValue(tftypes.Bool, v.Optional),
		})
	}
	return tftypes.NewValue(tftypes.Map{ElementType: configMapRefObjectType}, tfMap)
}

func ResourceFieldRefsToTfValue(refs map[string]ResourceFieldRef) tftypes.Value {
	tfMap := map[string]tftypes.Value{}
	for k, v := range refs {
		tfMap[k] = tftypes.NewValue(resourceFieldRefObjectType, map[string]tftypes.Value{
			"resource":  tftypes.NewValue(tftypes.String, v.Resource),
			"container": tftypes.NewValue(tftypes.String, v.Container),
			"divisor":   tftypes.NewValue(tftypes.String, v.Divisor),
		})
	}
	return tftypes.NewValue(tftypes.Map{ElementType: resourceFieldRefObjectType}, tfMap)
}

func FileKeyRefsToTfValue(refs map[string]FileKeyRef) tftypes.Value {
	tfMap := map[string]tftypes.Value{}
	for k, v := range refs {
		tfMap[k] = tftypes.NewValue(fileKeyRefObjectType, map[string]tftypes.Value{
			"key":         tftypes.NewValue(tftypes.String, v.Key),
			"path":        tftypes.NewValue(tftypes.String, v.Path),
			"volume_name": tftypes.NewValue(tftypes.String, v.VolumeName),
		})
	}
	return tftypes.NewValue(tftypes.Map{ElementType: fileKeyRefObjectType}, tfMap)
}

const envVariableKeyRegex = "^[a-zA-Z_][a-zA-Z0-9_]*$"

func validEnvVariableKey(key string) bool {
	regex := regexp.MustCompile(envVariableKeyRegex)
	return regex.MatchString(key)
}
