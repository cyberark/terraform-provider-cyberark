package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
)

func importedSyncPolicyState(t *testing.T, r *syncPolicyResource, id string) tfsdk.State {
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil)}
	diags := state.SetAttribute(context.Background(), path.Root("id"), types.StringValue(id))
	if diags.HasError() {
		t.Fatalf("setting the import state: %v", diags)
	}
	return state
}

func policyWithSafeFilter() *cybrapi.PolicyExternalOutput {
	return &cybrapi.PolicyExternalOutput{
		ID:     String("policy-1"),
		Source: &cybrapi.Source{SourceID: "store-source"},
		Target: &cybrapi.Target{TargetID: "store-target"},
		Filter: &cybrapi.FilterResponse{ID: String("filter-1"), Type: String("PAM_SAFE"), Data: &cybrapi.SafeDataFilter{SafeName: String("safe_1")}},
	}
}

func policyWithFilterIDOnly() *cybrapi.PolicyExternalOutput {
	policy := policyWithSafeFilter()
	policy.Filter = &cybrapi.FilterResponse{ID: String("filter-1")}
	return policy
}

func policyWithoutFilter() *cybrapi.PolicyExternalOutput {
	policy := policyWithSafeFilter()
	policy.Filter = nil
	return policy
}

func readImportedSyncPolicy(t *testing.T, shAPI *cybrapi.MockSecretsHubAPI) *resource.ReadResponse {
	r := &syncPolicyResource{api: &cybrapi.API{SecretsHubAPI: shAPI}}
	state := importedSyncPolicyState(t, r, "policy-1")
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)
	return resp
}

func assertSafeFilterInState(t *testing.T, resp *resource.ReadResponse) {
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	var got syncPolicyModel
	resp.Diagnostics.Append(resp.State.Get(context.Background(), &got)...)
	if got.SafeName.ValueString() != "safe_1" || got.Type.ValueString() != "PAM_SAFE" {
		t.Errorf("expected safe_name=safe_1 and safe_type=PAM_SAFE, got %q and %q", got.SafeName.ValueString(), got.Type.ValueString())
	}
}

func TestSyncPolicyResource_Read_SafeFilterInPolicyResponse(t *testing.T) {
	shAPI := cybrapi.NewMockSecretsHubAPI()
	shAPI.On("GetSyncPolicy", mock.Anything, "policy-1").Return(policyWithSafeFilter(), nil)

	resp := readImportedSyncPolicy(t, shAPI)

	assertSafeFilterInState(t, resp)
	shAPI.AssertNotCalled(t, "GetSecretFilter", mock.Anything, mock.Anything, mock.Anything)
}

func TestSyncPolicyResource_Read_FilterIDOnly_LooksUpTheFilter(t *testing.T) {
	shAPI := cybrapi.NewMockSecretsHubAPI()
	shAPI.On("GetSyncPolicy", mock.Anything, "policy-1").Return(policyWithFilterIDOnly(), nil)
	shAPI.On("GetSecretFilter", mock.Anything, "store-source", "filter-1").Return(&cybrapi.SecretFilterOutput{ID: String("filter-1"), Type: String("PAM_SAFE"), Data: &cybrapi.SafeDataFilter{SafeName: String("safe_1")}}, nil)

	resp := readImportedSyncPolicy(t, shAPI)

	assertSafeFilterInState(t, resp)
	shAPI.AssertCalled(t, "GetSecretFilter", mock.Anything, "store-source", "filter-1")
}

func TestSyncPolicyResource_Read_NoFilter_ReturnsFilterNotReadyError(t *testing.T) {
	shAPI := cybrapi.NewMockSecretsHubAPI()
	shAPI.On("GetSyncPolicy", mock.Anything, "policy-1").Return(policyWithoutFilter(), nil)

	resp := readImportedSyncPolicy(t, shAPI)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected an error diagnostic, got none")
	}
	err := resp.Diagnostics.Errors()[0]
	if err.Summary() != "Sync policy filter not ready" {
		t.Errorf("unexpected error summary %q", err.Summary())
	}
	if !strings.Contains(err.Detail(), "policy-1") {
		t.Errorf("expected the policy ID in the error detail, got %q", err.Detail())
	}
	shAPI.AssertNotCalled(t, "GetSecretFilter", mock.Anything, mock.Anything, mock.Anything)
}
