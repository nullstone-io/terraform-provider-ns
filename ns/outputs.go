package ns

import (
	"encoding/json"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
)

type Outputs map[string]Output

type Output struct {
	Type  *cty.Type       `json:"type"`
	Value json.RawMessage `json:"value"`
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

		rs := tfprotov5.RawState{JSON: output.Value}
		val, err := rs.Unmarshal(destType)
		if err != nil {
			return tftypes.Value{}, err
		}
		all[name] = val
	}
	return tftypes.NewValue(objType, all), nil
}
