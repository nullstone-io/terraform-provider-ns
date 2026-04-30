package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// Stubs for ProviderServer methods that this provider does not implement.
// These satisfy the tfprotov5.ProviderServer interface in newer SDK versions.

func (s *Server) GetMetadata(ctx context.Context, req *tfprotov5.GetMetadataRequest) (*tfprotov5.GetMetadataResponse, error) {
	resp := &tfprotov5.GetMetadataResponse{
		DataSources:        []tfprotov5.DataSourceMetadata{},
		Resources:          []tfprotov5.ResourceMetadata{},
		Functions:          []tfprotov5.FunctionMetadata{},
		EphemeralResources: []tfprotov5.EphemeralResourceMetadata{},
	}
	for typeName := range s.dsf {
		resp.DataSources = append(resp.DataSources, tfprotov5.DataSourceMetadata{TypeName: typeName})
	}
	for typeName := range s.rf {
		resp.Resources = append(resp.Resources, tfprotov5.ResourceMetadata{TypeName: typeName})
	}
	return resp, nil
}

func (s *Server) GetResourceIdentitySchemas(ctx context.Context, req *tfprotov5.GetResourceIdentitySchemasRequest) (*tfprotov5.GetResourceIdentitySchemasResponse, error) {
	return &tfprotov5.GetResourceIdentitySchemasResponse{
		IdentitySchemas: map[string]*tfprotov5.ResourceIdentitySchema{},
	}, nil
}

func (s *Server) MoveResourceState(ctx context.Context, req *tfprotov5.MoveResourceStateRequest) (*tfprotov5.MoveResourceStateResponse, error) {
	return &tfprotov5.MoveResourceStateResponse{
		Diagnostics: []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Move Resource State Not Supported",
				Detail:   "This provider does not support moving resource state.",
			},
		},
	}, nil
}

func (s *Server) UpgradeResourceIdentity(ctx context.Context, req *tfprotov5.UpgradeResourceIdentityRequest) (*tfprotov5.UpgradeResourceIdentityResponse, error) {
	return &tfprotov5.UpgradeResourceIdentityResponse{}, nil
}

func (s *Server) GenerateResourceConfig(ctx context.Context, req *tfprotov5.GenerateResourceConfigRequest) (*tfprotov5.GenerateResourceConfigResponse, error) {
	return &tfprotov5.GenerateResourceConfigResponse{
		Diagnostics: []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Generate Resource Config Not Supported",
				Detail:   "This provider does not support generating resource configurations.",
			},
		},
	}, nil
}

func (s *Server) CallFunction(ctx context.Context, req *tfprotov5.CallFunctionRequest) (*tfprotov5.CallFunctionResponse, error) {
	return &tfprotov5.CallFunctionResponse{
		Error: &tfprotov5.FunctionError{
			Text: "Functions are not supported by this provider.",
		},
	}, nil
}

func (s *Server) GetFunctions(ctx context.Context, req *tfprotov5.GetFunctionsRequest) (*tfprotov5.GetFunctionsResponse, error) {
	return &tfprotov5.GetFunctionsResponse{
		Functions: map[string]*tfprotov5.Function{},
	}, nil
}

func (s *Server) ValidateEphemeralResourceConfig(ctx context.Context, req *tfprotov5.ValidateEphemeralResourceConfigRequest) (*tfprotov5.ValidateEphemeralResourceConfigResponse, error) {
	return &tfprotov5.ValidateEphemeralResourceConfigResponse{
		Diagnostics: []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Ephemeral Resources Not Supported",
				Detail:   "This provider does not support ephemeral resources.",
			},
		},
	}, nil
}

func (s *Server) OpenEphemeralResource(ctx context.Context, req *tfprotov5.OpenEphemeralResourceRequest) (*tfprotov5.OpenEphemeralResourceResponse, error) {
	return &tfprotov5.OpenEphemeralResourceResponse{
		Diagnostics: []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Ephemeral Resources Not Supported",
				Detail:   "This provider does not support ephemeral resources.",
			},
		},
	}, nil
}

func (s *Server) RenewEphemeralResource(ctx context.Context, req *tfprotov5.RenewEphemeralResourceRequest) (*tfprotov5.RenewEphemeralResourceResponse, error) {
	return &tfprotov5.RenewEphemeralResourceResponse{
		Diagnostics: []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Ephemeral Resources Not Supported",
				Detail:   "This provider does not support ephemeral resources.",
			},
		},
	}, nil
}

func (s *Server) CloseEphemeralResource(ctx context.Context, req *tfprotov5.CloseEphemeralResourceRequest) (*tfprotov5.CloseEphemeralResourceResponse, error) {
	return &tfprotov5.CloseEphemeralResourceResponse{
		Diagnostics: []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Ephemeral Resources Not Supported",
				Detail:   "This provider does not support ephemeral resources.",
			},
		},
	}, nil
}
