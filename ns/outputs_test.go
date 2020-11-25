package ns

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputs_ToProtov5(t *testing.T) {
	two := 2
	tests := []struct {
		name      string
		inputFile string
		wantValue tftypes.Value
	}{
		{
			name:      "flat and homogeneous",
			inputFile: filepath.Join("test-fixtures", "state-files", "01.json"),
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
			name:      "flat and heterogeneous",
			inputFile: filepath.Join("test-fixtures", "state-files", "02.json"),
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
		{
			name:      "nested and heterogeneous",
			inputFile: filepath.Join("test-fixtures", "state-files", "03.json"),
			wantValue: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"key1": tftypes.String,
					"key2": tftypes.Number,
					"key3": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"key1": tftypes.String,
							"key2": tftypes.String,
							"key3": tftypes.String,
						},
					},
				},
			}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, "value1"),
				"key2": tftypes.NewValue(tftypes.Number, &two),
				"key3": tftypes.NewValue(tftypes.Object{
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
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := ioutil.ReadFile(test.inputFile)
			require.NoError(t, err, "read input file")
			var stateFile StateFile
			require.NoError(t, json.Unmarshal(raw, &stateFile), "unmarshal input file")

			gotValue, err := stateFile.Outputs.ToProtov5()
			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, test.wantValue, gotValue, "result")
		})
	}
}
