// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &awsSecretStoreResource{}
	_ resource.ResourceWithConfigure   = &awsSecretStoreResource{}
	_ resource.ResourceWithImportState = &awsSecretStoreResource{}
)

// NewAWSSecretStoreResource is a helper function to simplify the provider implementation.
func NewAWSSecretStoreResource() resource.Resource {
	return &awsSecretStoreResource{}
}

// awsSecretStoreResource defines the resource implementation.
type awsSecretStoreResource struct {
	api *cybrapi.API
}

// awsSecretStoreModel describes the resource data model.
type awsSecretStoreModel struct {
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	AccountAlias types.String `tfsdk:"aws_account_alias"`
	AccountID    types.String `tfsdk:"aws_account_id"`
	RegionID     types.String `tfsdk:"aws_account_region"`
	RoleName     types.String `tfsdk:"aws_iam_role"`
	ID           types.String `tfsdk:"id"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *awsSecretStoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_secret_store"
}

// Schema returns the resource schema.
func (r *awsSecretStoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `AWS Secret Store Resource

This resource is responsible for creating a new AWS secret store in Cyberark SecretsHub.

For more information click [here](https://docs.cyberark.com/secrets-hub-privilege-cloud/Latest/en/Content/Developer/sh-create-aws-target-tutorial.htm?tocpath=Developer%7CTutorials%7C_____1).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "CyberArk Privilege Cloud Secrets Store created from CyberArk after onboarding secret store into a secretshub.",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "Custom Secret Store Name for customizing the object name in a secret store.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description for target/secret store.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Should always be 'AWS_ASM' for AWS Secret Manager.",
				Computed:    true,
				Default:     stringdefault.StaticString("AWS_ASM"),
			},
			"aws_account_alias": schema.StringAttribute{
				Description: "AWS Account Alias ",
				Required:    true,
			},
			"aws_account_id": schema.StringAttribute{
				Description: "AWS Account ID",
				Required:    true,
			},
			"aws_account_region": schema.StringAttribute{
				Description: "AWS Region ID",
				Required:    true,
			},
			"aws_iam_role": schema.StringAttribute{
				Description: "AWS Role Name",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *awsSecretStoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create a new resource.
func (r *awsSecretStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data awsSecretStoreModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newStore := cybrapi.SecretStoreInput[cybrapi.AwsAsmData]{
		Name:        data.Name.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
		Type:        data.Type.ValueStringPointer(),
		Data: &cybrapi.AwsAsmData{
			AccountAlias: data.AccountAlias.ValueStringPointer(),
			AccountID:    data.AccountID.ValueStringPointer(),
			RegionID:     data.RegionID.ValueStringPointer(),
			RoleName:     data.RoleName.ValueStringPointer(),
		},
	}

	stores, err := r.api.SecretsHubAPI.GetAwsAsmSecretStores(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret stores",
			fmt.Sprintf("Error while reading secret stores: %+v", err))
		return
	}

	for _, store := range stores.SecretStores {
		if *store.Name == data.Name.ValueString() && *store.Data.AccountAlias == data.AccountAlias.ValueString() {
			tflog.Info(ctx, fmt.Sprintf("Secret store with name %s and account alias %s already exists", data.Name.ValueString(), data.AccountAlias.ValueString()))

			// We assume that secret store is already created
			data.ID = types.StringValue(store.ID)
			data.LastUpdated = types.StringPointerValue(store.UpdatedAt)
			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	output, err := r.api.SecretsHubAPI.AddAwsAsmSecretStore(ctx, newStore)
	if err != nil {
		resp.Diagnostics.AddError("Error creating secret store",
			fmt.Sprintf("Error while creating secret store: %+v", err))
		return
	}

	tflog.Info(ctx, "Secret Store created successfully")

	data.ID = types.StringValue(output.ID)
	data.LastUpdated = types.StringPointerValue(output.UpdatedAt)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read the resource state. The Read method is used to sync an existing resource with Terraform's state when Terraform is already aware of the resource.
func (r *awsSecretStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data awsSecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.api.SecretsHubAPI.GetAwsAsmSecretStore(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret store",
			fmt.Sprintf("Error while reading secret store: %+v", err))
		return
	}

	data = awsSecretStoreModel{
		Name:         types.StringPointerValue(output.Name),
		Description:  types.StringPointerValue(output.Description),
		Type:         types.StringPointerValue(output.Type),
		AccountAlias: types.StringPointerValue(output.Data.AccountAlias),
		AccountID:    types.StringPointerValue(output.Data.AccountID),
		RegionID:     types.StringPointerValue(output.Data.RegionID),
		RoleName:     types.StringPointerValue(output.Data.RoleName),
		ID:           types.StringValue(output.ID),
		LastUpdated:  types.StringPointerValue(output.UpdatedAt),
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *awsSecretStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state awsSecretStoreModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updatedStore := cybrapi.SecretStoreInput[cybrapi.AwsAsmData]{
		Name:        data.Name.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
		Type:        data.Type.ValueStringPointer(),
		Data: &cybrapi.AwsAsmData{
			AccountAlias: data.AccountAlias.ValueStringPointer(),
			AccountID:    data.AccountID.ValueStringPointer(),
			RegionID:     data.RegionID.ValueStringPointer(),
			RoleName:     data.RoleName.ValueStringPointer(),
		},
	}

	output, err := r.api.SecretsHubAPI.UpdateAwsSecretStore(ctx, state.ID.ValueString(), updatedStore)
	if err != nil {
		resp.Diagnostics.AddError("Error updating secret store",
			fmt.Sprintf("Error while updating secret store: %+v", err))
		return
	}

	data.ID = types.StringValue(output.ID)
	data.LastUpdated = types.StringPointerValue(output.UpdatedAt)

	tflog.Info(ctx, "AWS Secret Store updated successfully")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *awsSecretStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state awsSecretStoreModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.api.SecretsHubAPI.DeleteAwsSecretStore(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting AWS secret store",
			fmt.Sprintf("Error while deleting secret store: %+v", err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("AWS Secret Store %s deleted successfully", state.ID.ValueString()))
}

// ImportState imports an existing AWS Secret Store resource into Terraform. It retrieves the resource from CyberArk SecretsHub using the provided ID,
// sets the Terraform state to match the resource, and handles any errors.
func (r *awsSecretStoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
