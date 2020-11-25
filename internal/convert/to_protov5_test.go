package convert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToProtov5(t *testing.T) {
	two := 2
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantType  tftypes.Type
		wantValue tftypes.Value
	}{
		{
			name: "flat and homogeneous",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			wantType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"key1": tftypes.String,
					"key2": tftypes.String,
					"key3": tftypes.String,
				},
			},
			wantValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"key1": tftypes.String,
					"key2": tftypes.String,
					"key3": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, "value1"),
				"key2": tftypes.NewValue(tftypes.String, "value2"),
				"key3": tftypes.NewValue(tftypes.String, "value3"),
			}),
		},
		{
			name: "flat and heterogeneous",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 2,
				"key3": true,
			},
			wantType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"key1": tftypes.String,
					"key2": tftypes.Number,
					"key3": tftypes.Bool,
				},
			},
			wantValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"key1": tftypes.String,
					"key2": tftypes.Number,
					"key3": tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, "value1"),
				"key2": tftypes.NewValue(tftypes.Number, &two),
				"key3": tftypes.NewValue(tftypes.Bool, true),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotType, gotValue, err := ToProtov5(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.wantType, gotType)
			assert.Equal(t, test.wantValue, gotValue)
		})
	}
}
