package convert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

func ToProtov5(obj interface{}) (tftypes.Type, tftypes.Value, error) {
	switch x := obj.(type) {
	case bool:
		return tftypes.Bool, tftypes.NewValue(tftypes.Bool, x), nil
	case int:
		return tftypes.Number, tftypes.NewValue(tftypes.Number, &x), nil
	case float64:
		return tftypes.Number, tftypes.NewValue(tftypes.Number, &x), nil
	case string:
		return tftypes.String, tftypes.NewValue(tftypes.String, x), nil
	case []interface{}:
		return SliceToProtov5(x)
	case map[string]interface{}:
		return MapToProtov5(x)
	default:
		return nil, tftypes.NewValue(tftypes.String, nil), fmt.Errorf("unknown object type: %T", obj)
	}
}

func MapToProtov5(m map[string]interface{}) (tftypes.Type, tftypes.Value, error) {
	attrTypes := map[string]tftypes.Type{}
	val := map[string]tftypes.Value{}

	for key, child := range m {
		var err error
		attrTypes[key], val[key], err = ToProtov5(child)
		if err != nil {
			return nil, tftypes.NewValue(tftypes.String, nil), WrapConversionError(key, err)
		}
	}

	objType := tftypes.Object{AttributeTypes: attrTypes}
	return objType, tftypes.NewValue(objType, val), nil
}

func SliceToProtov5(slice []interface{}) (tftypes.Type, tftypes.Value, error) {
	val := make([]tftypes.Value, len(slice))
	for i, child := range slice {
		var err error
		_, val[i], err = ToProtov5(child)
		if err != nil {
			return nil, tftypes.NewValue(tftypes.String, nil), WrapConversionError(fmt.Sprintf("[%d]", i), err)
		}
	}

	elemType, err := tftypes.TypeFromElements(val)
	if err != nil {
		return nil, tftypes.NewValue(tftypes.String, nil), err
	}

	listType := tftypes.List{ElementType: elemType}
	return listType, tftypes.NewValue(listType, val), nil
}
