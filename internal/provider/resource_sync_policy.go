// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cybrapi "github.com/cyberark/terraform-provider-secretshub/internal/cyberark"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &syncPolicyResource{}
	_ resource.ResourceWithConfigure = &syncPolicyResource{}
)

// NewSyncPolicyResource is a helper function to simplify the provider implementation.
func NewSyncPolicyResource() resource.Resource {
	return &syncPolicyResource{}
}

// syncPolicyResource defines the resource implementation.
type syncPolicyResource struct {
	api *cybrapi.API
}

// syncPolicyModel describes the resource data model.
type syncPolicyModel struct {
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	SourceID       types.String `tfsdk:"source_id"`
	TargetID       types.String `tfsdk:"target_id"`
	Type           types.String `tfsdk:"safe_type"`
	SafeName       types.String `tfsdk:"safe_name"`
	Transformation types.String `tfsdk:"transformation"`
	ID             types.String `tfsdk:"id"`
	LastUpdated    types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *syncPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync_policy"
}

// Schema returns the resource schema.
func (r *syncPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Sync Policy Resource

This resource is responsible for creating a new sync policy to synchronize the secrets between cloud platforms (secret store) and Privilege Cloud using CyberArk Secrets Hub.

For more information click [here](https://docs.cyberark.com/secrets-hub-privilege-cloud/Latest/en/Content/Developer/sh-policy-api-tutorial.htm?tocpath=Developer%7CTutorials%7C_____4).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Sync policy Generated from CyberArk after onboarding policy into a secretshub.",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "Custom Sync Secret Store Policy Name for customizing the object name in a SecretsHub.",
				Required:    true,
			},
			"source_id": schema.StringAttribute{
				Description: "SourceID to sync secrets from",
				Required:    true,
			},
			"target_id": schema.StringAttribute{
				Description: "TargetID to sync secrets to",
				Required:    true,
			},
			"safe_type": schema.StringAttribute{
				Description: "Should always be PAM_SAFE for sync policy.",
				Computed:    true,
				Default:     stringdefault.StaticString("PAM_SAFE"),
			},
			"safe_name": schema.StringAttribute{
				Description: "Safe name need to be synced with target",
				Required:    true,
			},
			"transformation": schema.StringAttribute{
				Description: "To sync only the password as plain text to password-only-plain-text",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description for policy.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *syncPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*cybrapi.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected CreateAzureAkvData Source Configure Type",
			fmt.Sprintf("Expected *cybrapi.Api, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.api = api
}

// Create a new resource.
func (r *syncPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data syncPolicyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	sourceID := data.SourceID.ValueString()
	targetID := data.TargetID.ValueString()
	safeType := data.Type.ValueString()
	safeName := data.SafeName.ValueString()
	newPolicy := cybrapi.PolicyInput{
		Name: &name,
	}

	// Setting Optional values
	newPolicy.Transformation = transformationValue(data.Transformation)
	newPolicy.Description = data.Description.ValueStringPointer()

	filterDetails := cybrapi.Filter{
		Type: &safeType,
		Data: &cybrapi.SafeDataFilter{
			SafeName: &safeName,
		},
	}
	newPolicy.Source = &cybrapi.Source{
		SourceID: sourceID,
	}
	newPolicy.Target = &cybrapi.Target{
		TargetID: targetID,
	}
	newPolicy.Filter = &filterDetails

	policies, err := r.api.SecretsHubAPI.GetSyncPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read policies", fmt.Sprintf("Failed to read policies: %+v", err))
		return
	}

	var policy *cybrapi.PolicyExternalOutput

	for _, p := range policies.Policies {
		if *p.Name == name {
			policy = p
			break
		}
	}

	if policy == nil {
		tflog.Info(ctx, "Sync Policy not found, creating new")
		policy, err = r.api.SecretsHubAPI.AddSyncPolicy(ctx, newPolicy)
		if err != nil {
			resp.Diagnostics.AddError("Failed to create policy", fmt.Sprintf("Failed to create policy: %+v", err))
			return
		}
	}

	data.ID = types.StringPointerValue(policy.ID)
	data.LastUpdated = types.StringPointerValue(policy.UpdatedAt)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func transformationValue(transformation types.String) *cybrapi.TransformationValue {
	if transformation.IsNull() {
		return nil
	}
	return &cybrapi.TransformationValue{
		Predefined: transformation.ValueString(),
	}
}

// Read the resource state.
func (r *syncPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data syncPolicyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.api.SecretsHubAPI.GetSyncPolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read policy", fmt.Sprintf("Failed to read policy: %+v", err))
		return
	}

	data.ID = types.StringPointerValue(policy.ID)
	data.LastUpdated = types.StringPointerValue(policy.UpdatedAt)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported
func (r *syncPolicyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}

// Delete is not supported
func (r *syncPolicyResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}
