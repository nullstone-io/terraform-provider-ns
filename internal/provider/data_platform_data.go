package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nullstone-io/module/platformdata"
)

type dataPlatformData struct {
	p *provider
}

func newDataPlatformData(p *provider) (*dataPlatformData, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataPlatformData{p: p}, nil
}

func (*dataPlatformData) Schema(ctx context.Context) *tfprotov5.Schema {
	attrs := []*tfprotov5.SchemaAttribute{
		deprecatedIDAttribute(),
		{
			Name:            "kind",
			Type:            tftypes.String,
			Description:     "The kind of platform data (e.g. `env`). Determines the schema used to validate `data`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Required:        true,
		},
		{
			Name:            "version",
			Type:            tftypes.Number,
			Description:     "The schema version of `kind` that `data` conforms to. Must be a positive integer.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Required:        true,
		},
		{
			Name:            "data",
			Type:            tftypes.String,
			Description:     "A JSON-encoded object (typically produced with `jsonencode(...)`) that conforms to the schema for `kind`/`version`.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Required:        true,
		},
	}

	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description: "No-op data source that persists platform data into Terraform state. " +
				"Platform data is consumed only by Nullstone (UI, API, CLI, deployment tooling) and is never read by other modules. " +
				"This data source performs no API calls; it validates the payload and echoes it into state.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes:      attrs,
		},
	}
}

func (d *dataPlatformData) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	_, diags := validatePlatformData(config)
	return diags, nil
}

func (d *dataPlatformData) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	envelope, diags := validatePlatformData(config)
	if hasErrorDiagnostic(diags) {
		return nil, diags, nil
	}

	return map[string]tftypes.Value{
		"id":      tftypes.NewValue(tftypes.String, fmt.Sprintf("%s/%d", envelope.Kind, envelope.Version)),
		"kind":    config["kind"],
		"version": config["version"],
		"data":    config["data"],
	}, diags, nil
}

// validatePlatformData builds the platform data envelope from config and validates it.
// If any input is null or not fully known (e.g. derived from an unknown expression during plan),
// validation is skipped and no diagnostics are returned; Terraform calls Read once values are known.
func validatePlatformData(config map[string]tftypes.Value) (platformdata.Envelope, []*tfprotov5.Diagnostic) {
	kindVal, versionVal, dataVal := config["kind"], config["version"], config["data"]
	for _, v := range []tftypes.Value{kindVal, versionVal, dataVal} {
		if v.IsNull() || !v.IsFullyKnown() {
			return platformdata.Envelope{}, nil
		}
	}

	var kind string
	if err := kindVal.As(&kind); err != nil {
		return platformdata.Envelope{}, []*tfprotov5.Diagnostic{invalidPlatformDataDiag(fmt.Sprintf("kind: %s", err))}
	}

	versionFloat := new(big.Float)
	if err := versionVal.As(&versionFloat); err != nil {
		return platformdata.Envelope{}, []*tfprotov5.Diagnostic{invalidPlatformDataDiag(fmt.Sprintf("version: %s", err))}
	}
	if !versionFloat.IsInt() {
		return platformdata.Envelope{}, []*tfprotov5.Diagnostic{invalidPlatformDataDiag(fmt.Sprintf("version must be a positive integer, got %s", versionFloat.Text('f', -1)))}
	}
	version, _ := versionFloat.Int64()

	var data string
	if err := dataVal.As(&data); err != nil {
		return platformdata.Envelope{}, []*tfprotov5.Diagnostic{invalidPlatformDataDiag(fmt.Sprintf("data: %s", err))}
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &probe); err != nil || probe == nil {
		return platformdata.Envelope{}, []*tfprotov5.Diagnostic{invalidPlatformDataDiag("data must be a JSON-encoded object (use `jsonencode({...})`)")}
	}

	envelope := platformdata.Envelope{
		Kind:    kind,
		Version: int(version),
		Data:    json.RawMessage(data),
	}
	warnings, err := platformdata.ValidateEnvelope(envelope)
	if err != nil {
		return envelope, []*tfprotov5.Diagnostic{invalidPlatformDataDiag(err.Error())}
	}
	diags := make([]*tfprotov5.Diagnostic, 0, len(warnings))
	for _, warning := range warnings {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityWarning,
			Summary:  "Unrecognized platform data",
			Detail:   warning,
		})
	}
	return envelope, diags
}

func invalidPlatformDataDiag(detail string) *tfprotov5.Diagnostic {
	return &tfprotov5.Diagnostic{
		Severity: tfprotov5.DiagnosticSeverityError,
		Summary:  "Invalid platform data",
		Detail:   detail,
	}
}

func hasErrorDiagnostic(diags []*tfprotov5.Diagnostic) bool {
	for _, diag := range diags {
		if diag != nil && diag.Severity == tfprotov5.DiagnosticSeverityError {
			return true
		}
	}
	return false
}
