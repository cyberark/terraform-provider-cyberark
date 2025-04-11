// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &syncPolicyResource{}
	_ resource.ResourceWithConfigure      = &syncPolicyResource{}
	_ resource.ResourceWithValidateConfig = &syncPolicyResource{}
	_ resource.ResourceWithImportState    = &syncPolicyResource{}
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
				Description: "To sync only the password as plain text to password_only_plain_text",
				Computed:    true,
				Default:     stringdefault.StaticString("password_only_plain_text"),
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
			"Unexpected AzureAkvData Source Configure Type",
			fmt.Sprintf("Expected *cybrapi.Api, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.api = api
}

// ValidateConfig validates the resource configuration.
func (r *syncPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data syncPolicyModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the transformation value if provided
	if !data.Transformation.IsNull() && data.Transformation != types.StringValue("default") && data.Transformation != types.StringValue("password_only_plain_text") {
		resp.Diagnostics.AddError("Invalid Transformation Value",
			fmt.Sprintf("Transformation value must be either 'default' or 'password_only_plain_text', got: %s", data.Transformation.ValueString()),
		)
	}
}

// Create a new resource.
func (r *syncPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data syncPolicyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newPolicy := cybrapi.PolicyInput{
		Name:        data.Name.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
		Source: &cybrapi.Source{
			SourceID: data.SourceID.ValueString(),
		},
		Target: &cybrapi.Target{
			TargetID: data.TargetID.ValueString(),
		},
		Filter: &cybrapi.Filter{
			Type: data.Type.ValueStringPointer(),
			Data: &cybrapi.SafeDataFilter{
				SafeName: data.SafeName.ValueStringPointer(),
			},
		},
		Transformation: &cybrapi.TransformationValue{
			Predefined: data.Transformation.ValueString(),
		},
	}

	policies, err := r.api.SecretsHubAPI.GetSyncPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read policies", fmt.Sprintf("Failed to read policies: %+v", err))
		return
	}

	var policy *cybrapi.PolicyExternalOutput

	for _, p := range policies.Policies {
		if *p.Name == data.Name.ValueString() {
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

	store, err := r.api.SecretsHubAPI.GetSecretFilter(ctx, policy.Source.SourceID, *policy.Filter.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read store ", fmt.Sprintf("Failed to read store: %+v", err))
		return
	}

	data = syncPolicyModel{
		Name:        types.StringPointerValue(policy.Name),
		Description: types.StringPointerValue(policy.Description),
		SourceID:    types.StringValue(policy.Source.SourceID),
		TargetID:    types.StringValue(policy.Target.TargetID),
		Type:        types.StringPointerValue(store.Type),
		SafeName:    types.StringPointerValue(store.Data.SafeName),
		ID:          types.StringPointerValue(policy.ID),
		LastUpdated: types.StringPointerValue(policy.UpdatedAt),
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported for this resource.
func (r *syncPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state syncPolicyModel

	// Read Terraform plan data and current state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the new policy configuration
	updatePolicy := cybrapi.PolicyInput{
		Name:        data.Name.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
		Source: &cybrapi.Source{
			SourceID: data.SourceID.ValueString(),
		},
		Target: &cybrapi.Target{
			TargetID: data.TargetID.ValueString(),
		},
		Filter: &cybrapi.Filter{
			Type: data.Type.ValueStringPointer(),
			Data: &cybrapi.SafeDataFilter{
				SafeName: data.SafeName.ValueStringPointer(),
			},
		},
		Transformation: &cybrapi.TransformationValue{
			Predefined: data.Transformation.ValueString(),
		},
	}

	// Call API to update (delete and recreate) the policy
	policy, err := r.api.SecretsHubAPI.UpdateSyncPolicy(ctx, state.ID.ValueString(), updatePolicy)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update sync policy",
			fmt.Sprintf("Failed to update sync policy: %+v", err))
		return
	}

	// Update the state with the new policy information
	data.ID = types.StringPointerValue(policy.ID)
	data.LastUpdated = types.StringPointerValue(policy.UpdatedAt)

	tflog.Info(ctx, "Sync Policy updated successfully")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete removes the resource and deletes the Terraform state on success.
func (r *syncPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state syncPolicyModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call the API to delete the policy
	err := r.api.SecretsHubAPI.DeleteSyncPolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete sync policy",
			fmt.Sprintf("Failed to delete sync policy: %+v", err))
		return
	}

	tflog.Info(ctx, "Sync Policy deleted successfully")
}

func (r *syncPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
