package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type dataAppEnv struct {
	p *provider
}

func newDataAppEnv(p *provider) (*dataAppEnv, error) {
	if p == nil {
		return nil, fmt.Errorf("a provider is required")
	}
	return &dataAppEnv{p: p}, nil
}

func (*dataAppEnv) Schema(ctx context.Context) *tfprotov5.Schema {
	attrs := []*tfprotov5.SchemaAttribute{
		deprecatedIDAttribute(),
		{
			Name:            "stack_id",
			Type:            tftypes.Number,
			Required:        true,
			Description:     "The ID of the owning stack for the application in nullstone.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
		},
		{
			Name:            "app_id",
			Type:            tftypes.Number,
			Required:        true,
			Description:     "The ID of the application in nullstone.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
		},
		{
			Name:            "env_id",
			Type:            tftypes.Number,
			Required:        true,
			Description:     "The ID of the environment in nullstone.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
		},
		{
			Name:            "version",
			Type:            tftypes.String,
			Computed:        true,
			Description:     "The version configured for the application in the specific environment.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
		},
	}

	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to configure module based on current nullstone workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes:      attrs,
		},
	}
}

func (d *dataAppEnv) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataAppEnv) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	stackId := extractInt64FromConfig(config, "stack_id")
	appId := extractInt64FromConfig(config, "app_id")
	envId := extractInt64FromConfig(config, "env_id")
	diags := make([]*tfprotov5.Diagnostic, 0)

	var appEnvId int64
	var appEnvVersion string

	nsClient := api.Client{Config: d.p.NsConfig}
	app, err := d.findApp(stackId, appId)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("Unable to list applications."),
			Detail:   err.Error(),
		})
	} else if app == nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The application environment (stack=%d, app=%d, env=%d) is missing.", stackId, appId, envId),
		})
	} else if env, err := d.findEnv(stackId, envId); err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("Unable to find environment (stack=%d, env=%d).", stackId, envId),
		})
	} else if env == nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The environment (stack=%d, env=%d) is missing.", stackId, envId),
		})
	} else {
		appEnv, err := nsClient.AppEnvs().Get(app.Id, env.Name)
		if err != nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Unable to retrieve the application environment (stack=%d, app=%d, env=%d) is missing.", stackId, appId, envId),
				Detail:   err.Error(),
			})
		} else if appEnv == nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Unable to find the application environment (stack=%d, app=%d, env=%d) is missing.", stackId, appId, envId),
				Detail:   err.Error(),
			})
		} else {
			appEnvId = appEnv.Id
			appEnvVersion = appEnv.Version
		}
	}

	return map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, fmt.Sprintf("%d", appEnvId)),
		"app_id":   tftypes.NewValue(tftypes.Number, &appId),
		"stack_id": tftypes.NewValue(tftypes.Number, &stackId),
		"env_id":   tftypes.NewValue(tftypes.Number, &envId),
		"version":  tftypes.NewValue(tftypes.String, appEnvVersion),
	}, diags, nil
}

func (d *dataAppEnv) findApp(stackId, appId int64) (*types.Application, error) {
	nsClient := api.Client{Config: d.p.NsConfig}
	apps, err := nsClient.Apps().List()
	if err != nil {
		return nil, fmt.Errorf("Unable to list applications.")
	}
	for _, app := range apps {
		if app.Id == appId && app.StackId == stackId {
			return &app, nil
		}
	}
	return nil, nil
}

func (d *dataAppEnv) findEnv(stackId, envId int64) (*types.Environment, error) {
	nsClient := api.Client{Config: d.p.NsConfig}
	return nsClient.Environments().Get(stackId, envId)
}
