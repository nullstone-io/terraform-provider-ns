package ns

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

// TFtypeFromCtyType is lifted from github.com/hashicorp/terraform-plugin-sdk/internal/plugin/convert
// It's very odd that this isn't exported because it was very valuable to parsing outputs
func TftypeFromCtyType(in cty.Type) (tftypes.Type, error) {
	switch {
	case in.Equals(cty.String):
		return tftypes.String, nil
	case in.Equals(cty.Number):
		return tftypes.Number, nil
	case in.Equals(cty.Bool):
		return tftypes.Bool, nil
	case in.Equals(cty.DynamicPseudoType):
		return tftypes.DynamicPseudoType, nil
	case in.IsSetType():
		elemType, err := TftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.Set{
			ElementType: elemType,
		}, nil
	case in.IsListType():
		elemType, err := TftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.List{
			ElementType: elemType,
		}, nil
	case in.IsTupleType():
		elemTypes := make([]tftypes.Type, 0, in.Length())
		for _, typ := range in.TupleElementTypes() {
			elemType, err := TftypeFromCtyType(typ)
			if err != nil {
				return nil, err
			}
			elemTypes = append(elemTypes, elemType)
		}
		return tftypes.Tuple{
			ElementTypes: elemTypes,
		}, nil
	case in.IsMapType():
		elemType, err := TftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.Map{
			AttributeType: elemType,
		}, nil
	case in.IsObjectType():
		attrTypes := make(map[string]tftypes.Type)
		for key, typ := range in.AttributeTypes() {
			attrType, err := TftypeFromCtyType(typ)
			if err != nil {
				return nil, err
			}
			attrTypes[key] = attrType
		}
		return tftypes.Object{
			AttributeTypes: attrTypes,
		}, nil
	}
	return nil, fmt.Errorf("unknown cty type %s", in.GoString())
}
