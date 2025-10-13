package provider

import (
	"context"
	"strings"
	"testing"
	"time"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSafeResource_Metadata(t *testing.T) {
	r := &pvwaSafeResource{}
	req := resource.MetadataRequest{ProviderTypeName: "cyberark"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)
	if resp.TypeName != "cyberark_pvwa_safe" {
		t.Errorf("expected TypeName to be 'cyberark_safe', got '%s'", resp.TypeName)
	}
}

func TestSafeResource_Schema_AttributeProperties(t *testing.T) {
	r := &pvwaSafeResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	tests := []struct {
		name     string
		required bool
		optional bool
		computed bool
	}{
		{"id", false, false, true},
		{"id_number", false, false, true},
		{"last_updated", false, false, true},
		{"safe_name", true, false, false},
		{"member", true, false, false},
		{"member_type", true, false, false},
		{"permission_level", true, false, false},
		{"safe_desc", false, true, true},
		{"safe_loc", false, true, true},
		{"cpm_name", false, true, true},
		{"retention", false, true, true},
		{"retention_versions", false, true, false},
		{"purge", false, true, true},
	}

	for _, tt := range tests {
		attr, ok := resp.Schema.Attributes[tt.name]
		if !ok {
			t.Errorf("attribute %q not found in schema", tt.name)
			continue
		}
		if attr.IsRequired() != tt.required {
			t.Errorf("attribute %q: expected Required=%v, got %v", tt.name, tt.required, attr.IsRequired())
		}
		if attr.IsOptional() != tt.optional {
			t.Errorf("attribute %q: expected Optional=%v, got %v", tt.name, tt.optional, attr.IsOptional())
		}
		if attr.IsComputed() != tt.computed {
			t.Errorf("attribute %q: expected Computed=%v, got %v", tt.name, tt.computed, attr.IsComputed())
		}
	}
}

func TestValidateSafeModel(t *testing.T) {
	tests := []struct {
		name       string
		model      pvwaSafeResourceModel
		wantErrors []string // substrings to expect in diagnostics
	}{
		{
			name: "valid - full permission, no retention fields",
			model: pvwaSafeResourceModel{
				PermType:          types.StringValue("full"),
				RetentionDays:     types.Int64PointerValue(nil),
				RetentionVersions: types.Int64PointerValue(nil),
			},
			wantErrors: nil,
		},
		{
			name: "invalid permission",
			model: pvwaSafeResourceModel{
				PermType:          types.StringValue("invalid"),
				RetentionDays:     types.Int64PointerValue(nil),
				RetentionVersions: types.Int64PointerValue(nil),
			},
			wantErrors: []string{"Permission Level Error", "invalid"},
		},
		{
			name: "both retention fields set",
			model: pvwaSafeResourceModel{
				PermType:          types.StringValue("read"),
				RetentionDays:     types.Int64Value(7),
				RetentionVersions: types.Int64Value(5),
			},
			wantErrors: []string{"Retention Fields Error", "only one of 'retention'"},
		},
		{
			name: "both validators fail",
			model: pvwaSafeResourceModel{
				PermType:          types.StringValue("nope"),
				RetentionDays:     types.Int64Value(3),
				RetentionVersions: types.Int64Value(4),
			},
			wantErrors: []string{"Permission Level Error", "Retention Fields Error"},
		},
		{
			name: "valid permission, retentionDays only",
			model: pvwaSafeResourceModel{
				PermType:          types.StringValue("manager"),
				RetentionDays:     types.Int64Value(15),
				RetentionVersions: types.Int64PointerValue(nil),
			},
			wantErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateSafeModel(tt.model)

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

func TestSafeResource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
		wantError    bool
		errorMessage string
	}{
		{
			name:         "nil_provider_data",
			providerData: nil,
			wantError:    false,
		},
		{
			name:         "valid_api_client",
			providerData: &cybrapi.API{},
			wantError:    false,
		},
		{
			name:         "invalid_provider_data_type",
			providerData: "invalid",
			wantError:    true,
			errorMessage: "Expected *cybrapi.Api, got: string",
		},
		{
			name:         "invalid_provider_data_struct",
			providerData: struct{}{},
			wantError:    true,
			errorMessage: "Expected *cybrapi.Api, got: struct {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &pvwaSafeResource{}
			req := resource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &resource.ConfigureResponse{}

			r.Configure(context.Background(), req, resp)

			if tt.wantError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("expected error, got none")
				}
				if len(resp.Diagnostics.Errors()) > 0 {
					errorMsg := resp.Diagnostics.Errors()[0].Summary()
					if errorMsg != "Unexpected Resource Configure Type" {
						t.Errorf("expected error summary 'Unexpected Resource Configure Type', got '%s'", errorMsg)
					}
					detailMsg := resp.Diagnostics.Errors()[0].Detail()
					if !strings.Contains(detailMsg, tt.errorMessage) {
						t.Errorf("expected error detail to contain '%s', got '%s'", tt.errorMessage, detailMsg)
					}
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("unexpected error: %v", resp.Diagnostics.Errors())
				}
				if tt.providerData != nil {
					if r.api == nil {
						t.Error("expected api to be set, got nil")
					}
				}
			}
		})
	}
}

func TestGetLastUpdated(t *testing.T) {
	// Case 1: lastModificationTime is not nil
	now := time.Now()
	micro := now.UnixMicro()
	got := getLastUpdated(&micro)
	want := types.StringValue(time.UnixMicro(micro).Format(time.RFC3339))
	if got.ValueString() != want.ValueString() {
		t.Errorf("expected %s, got %s", want.ValueString(), got.ValueString())
	}

	// Case 2: lastModificationTime is nil
	gotNil := getLastUpdated(nil)
	// Allow a small time difference due to execution time
	gotTime, err := time.Parse(time.RFC3339, gotNil.ValueString())
	if err != nil {
		t.Fatalf("invalid RFC3339 time: %v", err)
	}
	if time.Since(gotTime) > time.Second {
		t.Errorf("expected time within 1s of now, got %s", gotNil.ValueString())
	}
}
