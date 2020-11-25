package ns

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type Outputs map[string]Output

type Output struct {
	Type  *cty.Type   `json:"type"`
	Value interface{} `json:"value"`
}

func (o Outputs) ToProtov5() (tftypes.Value, error) {
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}
	all := map[string]tftypes.Value{}

	for name, output := range o {
		destType, err := TftypeFromCtyType(*output.Type)
		if err != nil {
			return tftypes.Value{}, err
		}
		objType.AttributeTypes[name] = destType
		all[name] = valToProtov5(destType, output.Value)
	}
	return tftypes.NewValue(objType, all), nil
}

func valToProtov5(destType tftypes.Type, value interface{}) tftypes.Value {
	// Normally, we wouldn't need this switch statement
	// However, tftypes.NewValue doesn't accept a value of type float64 or int
	// It appears like an oversight that they only accept *int and *float64
	// We're going to force that type for value as a way to make this happy
	//switch x := value.(type) {
	//case int:
	//	value = &x
	//case float64:
	//	value = &x
	//}

	switch {
	case destType.Is(tftypes.Set{}):
		return setOrListToProtov5(destType, destType.(tftypes.Set).ElementType, value.([]interface{}))
	case destType.Is(tftypes.List{}):
		return setOrListToProtov5(destType, destType.(tftypes.List).ElementType, value.([]interface{}))
	case destType.Is(tftypes.Tuple{}):
		return tupleToProtov5(destType.(tftypes.Tuple), value.([]interface{}))
	case destType.Is(tftypes.Map{}):
		return mapToProtov5(destType.(tftypes.Map), value.(map[string]interface{}))
	case destType.Is(tftypes.Object{}):
		return objectToProtov5(destType.(tftypes.Object), value.(map[string]interface{}))
	default: // String, Number, Bool, DynamicPseudoType
		return tftypes.NewValue(destType, value)
	}
}

func setOrListToProtov5(destType tftypes.Type, elemType tftypes.Type, s []interface{}) tftypes.Value {
	all := make([]tftypes.Value, len(s))
	for i, child := range s {
		all[i] = valToProtov5(elemType, child)
	}
	return tftypes.NewValue(destType, all)
}

func tupleToProtov5(destType tftypes.Tuple, s []interface{}) tftypes.Value {
	all := make([]tftypes.Value, len(s))
	for i, child := range s {
		all[i] = valToProtov5(destType.ElementTypes[i], child)
	}
	return tftypes.NewValue(destType, all)
}

func mapToProtov5(destType tftypes.Map, m map[string]interface{}) tftypes.Value {
	all := map[string]tftypes.Value{}
	for name, child := range m {
		all[name] = valToProtov5(destType.AttributeType, child)
	}
	return tftypes.NewValue(destType, all)
}

func objectToProtov5(destType tftypes.Object, m map[string]interface{}) tftypes.Value {
	all := map[string]tftypes.Value{}
	for name, child := range m {
		all[name] = valToProtov5(destType.AttributeTypes[name], child)
	}
	return tftypes.NewValue(destType, all)
}
