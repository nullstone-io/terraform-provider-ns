package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"strconv"
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
	return &tfprotov5.Schema{
		Version: 1,
		Block: &tfprotov5.SchemaBlock{
			Description:     "Data source to configure module based on current nullstone workspace.",
			DescriptionKind: tfprotov5.StringKindMarkdown,
			Attributes: []*tfprotov5.SchemaAttribute{
				deprecatedIDAttribute(),
				{
					Name:            "app",
					Type:            tftypes.String,
					Required:        true,
					Description:     "The name of the application in nullstone.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "stack",
					Type:            tftypes.String,
					Required:        true,
					Description:     "The owning stack for the application in nullstone.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "env",
					Type:            tftypes.String,
					Required:        true,
					Description:     "The name of the environment in nullstone.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
				{
					Name:            "version",
					Type:            tftypes.String,
					Computed:        true,
					Description:     "The version configured for the application in the specific environment.",
					DescriptionKind: tfprotov5.StringKindMarkdown,
				},
			},
		},
	}
}

func (d *dataAppEnv) Validate(ctx context.Context, config map[string]tftypes.Value) ([]*tfprotov5.Diagnostic, error) {
	return nil, nil
}

func (d *dataAppEnv) Read(ctx context.Context, config map[string]tftypes.Value) (map[string]tftypes.Value, []*tfprotov5.Diagnostic, error) {
	appName := extractStringFromConfig(config, "app")
	stackName := extractStringFromConfig(config, "stack")
	envName := extractStringFromConfig(config, "env")
	diags := make([]*tfprotov5.Diagnostic, 0)

	var appEnvId int
	var appEnvVersion string

	nsClient := api.Client{Config: d.p.NsConfig}
	app, err := d.findApp(appName, stackName)
	if err != nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("Unable to list applications."),
			Detail:   err.Error(),
		})
	} else if app == nil {
		diags = append(diags, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  fmt.Sprintf("The application environment (app=%s, stack=%s, env=%s) is missing.", appName, stackName, envName),
		})
	} else {
		appEnv, err := nsClient.AppEnvs().Get(app.Id, envName)
		if err != nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Unable to retrieve the application environment (app=%s, stack=%s, env=%s).", appName, stackName, envName),
				Detail:   err.Error(),
			})
		} else if appEnv == nil {
			diags = append(diags, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  fmt.Sprintf("Unable to find the application environment (app=%s, stack=%s, env=%s).", appName, stackName, envName),
				Detail:   err.Error(),
			})
		} else {
			appEnvId = appEnv.Id
			appEnvVersion = appEnv.Version
		}
	}

	return map[string]tftypes.Value{
		"id":      tftypes.NewValue(tftypes.String, strconv.Itoa(appEnvId)),
		"app":     tftypes.NewValue(tftypes.String, appName),
		"stack":   tftypes.NewValue(tftypes.String, stackName),
		"env":     tftypes.NewValue(tftypes.String, envName),
		"version": tftypes.NewValue(tftypes.String, appEnvVersion),
	}, diags, nil
}

func (d *dataAppEnv) findApp(appName, stackName string) (*types.Application, error) {
	nsClient := api.Client{Config: d.p.NsConfig}
	apps, err := nsClient.Apps().List()
	if err != nil {
		return nil, fmt.Errorf("Unable to list applications.")
	}
	for _, app := range apps {
		if app.Name == appName && app.StackName == stackName {
			return &app, nil
		}
	}
	return nil, nil
}
