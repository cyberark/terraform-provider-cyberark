package provider

import (
	"context"
	"strings"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwprovider.SchemaRequest{}
	schemaResponse := &fwprovider.SchemaResponse{}

	// Instantiate the provider.Provider and call its Schema method
	New("test")().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestValidateProviderConfigModel(t *testing.T) {
	tests := []struct {
		name       string
		model      secretsHubProviderModel
		wantErrors []string // substrings to expect in diagnostics
	}{
		{
			name: "valid cloud config",
			model: secretsHubProviderModel{
				Tenant:       types.StringValue("tenant"),
				ClientID:     types.StringValue("id"),
				ClientSecret: types.StringValue("secret"),
				Domain:       types.StringValue("domain"),
			},
			wantErrors: nil,
		},
		{
			name: "valid pvwa config",
			model: secretsHubProviderModel{
				PVWAUsername:    types.StringValue("user"),
				PVWAPassword:    types.StringValue("pass"),
				PVWAURL:         types.StringValue("https://pvwa"),
				PVWALoginMethod: types.StringValue("cyberark"),
			},
			wantErrors: nil,
		},
		{
			name: "missing required cloud attribute",
			model: secretsHubProviderModel{
				Tenant:       types.StringValue("tenant"),
				ClientID:     types.StringValue("id"),
				ClientSecret: types.StringValue(""),
				Domain:       types.StringValue("domain"),
			},
			wantErrors: []string{"Missing required attribute: client_secret"},
		},
		{
			name: "invalid pvwa login method",
			model: secretsHubProviderModel{
				PVWAUsername:    types.StringValue("user"),
				PVWAPassword:    types.StringValue("pass"),
				PVWAURL:         types.StringValue("https://pvwa"),
				PVWALoginMethod: types.StringValue("bad"),
			},
			wantErrors: []string{"login method (bad) is invalid"},
		},
		{
			name: "missing required pvwa attributes",
			model: secretsHubProviderModel{
				PVWAUsername: types.StringValue("user"),
				PVWAPassword: types.StringValue("pass"),
			},
			wantErrors: []string{"Missing required attribute: pvwa_url"},
		},
		{
			name:       "missing all config",
			model:      secretsHubProviderModel{},
			wantErrors: []string{"At least one of the following sets of attributes must be provided"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateProviderConfigModel(tt.model)

			if len(tt.wantErrors) == 0 && diags.HasError() {
				t.Fatalf("unexpected diagnostics: %#v", diags)
			}

			for _, want := range tt.wantErrors {
				matched := false
				for _, d := range diags {
					if strings.Contains(d.Summary(), want) || strings.Contains(d.Detail(), want) {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("expected diagnostic containing %q, got: %#v", want, diags)
				}
			}
		})
	}
}
