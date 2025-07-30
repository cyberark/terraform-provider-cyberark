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
	_ resource.Resource                   = &gcpSecretStoreResource{}
	_ resource.ResourceWithConfigure      = &gcpSecretStoreResource{}
	_ resource.ResourceWithValidateConfig = &gcpSecretStoreResource{}
	_ resource.ResourceWithImportState    = &gcpSecretStoreResource{}
)

// NewGcpSecretStoreResource is a helper function to simplify the provider implementation.
func NewGcpSecretStoreResource() resource.Resource {
	return &gcpSecretStoreResource{}
}

// gcpSecretStoreResource defines the resource implementation.
type gcpSecretStoreResource struct {
	api *cybrapi.API
}

// gcpSecretStoreModel describes the resource data model.
type gcpSecretStoreModel struct {
	Name                         types.String `tfsdk:"name"`
	Description                  types.String `tfsdk:"description"`
	Type                         types.String `tfsdk:"type"`
	GcpProjectName               types.String `tfsdk:"gcp_project_name"`
	GcpProjectNumber             types.String `tfsdk:"gcp_project_number"`
	GcpWorkloadIdentityPoolId    types.String `tfsdk:"gcp_workload_identity_pool_id"`
	GcpPoolProviderId            types.String `tfsdk:"gcp_pool_provider_id"`
	ServiceAccountEmail          types.String `tfsdk:"service_account_email"`
	ID                           types.String `tfsdk:"id"`
	LastUpdated                  types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *gcpSecretStoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_secret_store"
}

// Schema returns the resource schema.
func (r *gcpSecretStoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Gcp Secret Store Resource

This resource is responsible for creating and managing a GCP Secret Store in CyberArk SecretsHub.
It supports full CRUD (Create, Read, Update, Delete) operations and allows for the import of existing secret store configurations.

For more information, visit the CyberArk documentation.`,
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
				Description: "Should always be 'GCP_GSM'.",
				Computed:    true,
				Default:     stringdefault.StaticString("GCP_GSM"),
			},
			"gcp_project_name": schema.StringAttribute{
				Description: "GCP Project Name.",
				Required:    true,
			},
			"gcp_project_number": schema.StringAttribute{
				Description: "GCP Project Number.",
				Required:    true,
			},
			"gcp_workload_identity_pool_id": schema.StringAttribute{
				Description: "GCP Workload Identity Pool ID.",
				Required:    true,
			},
			"gcp_pool_provider_id": schema.StringAttribute{
				Description: "GCP Pool Provider ID.",
				Required:    true,
			},
			"service_account_email": schema.StringAttribute{
				Description: "Service Account Email.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *gcpSecretStoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*cybrapi.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Gcp Data Source Configure Type",
			fmt.Sprintf("Expected *cybrapi.Api, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.api = api
}

// ValidateConfig validates the resource configuration.
func (r *gcpSecretStoreResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data gcpSecretStoreModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the InputFields
	if err := cybrapi.ValidateInputField(ctx, "name", data.Name, 1, 200, "^[a-zA-Z0-9!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~]+$"); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
	}
	if err := cybrapi.ValidateInputField(ctx, "description", data.Description, 1, 150, `^[A-Za-z0-9-_,.();: ]+$`); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
	}
	if err := cybrapi.ValidateInputField(ctx, "gcp_project_name", data.GcpProjectName, 4, 30, `^[a-zA-Z0-9'"! -]+$`); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
	}
	if err := cybrapi.ValidateInputField(ctx, "gcp_project_number", data.GcpProjectNumber, 1, 18, `^[0-9]+$`); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
	}
	if err := cybrapi.ValidateInputField(ctx, "gcp_workload_identity_pool_id", data.GcpWorkloadIdentityPoolId, 4, 32, `^[a-z0-9-]+$`); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
	}
	if err := cybrapi.ValidateInputField(ctx, "gcp_pool_provider_id", data.GcpPoolProviderId, 4, 32, `^[a-z0-9-]+$`); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
	}
	if err := cybrapi.ValidateInputField(ctx, "service_account_email", data.ServiceAccountEmail, 37, 86, `^[a-z0-9-]{6,30}@[a-z0-9-]{6,30}\.iam\.gserviceaccount\.com$`); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
	}
}

// Create a new resource.
func (r *gcpSecretStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data gcpSecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newStore := cybrapi.SecretStoreInput[cybrapi.GcpData]{
		Name:        data.Name.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
		Type:        data.Type.ValueStringPointer(),
		Data: &cybrapi.GcpData{
			GcpProjectName:  data.GcpProjectName.ValueStringPointer(),
			GcpProjectNumber:  data.GcpProjectNumber.ValueStringPointer(),
			GcpWorkloadIdentityPoolId:  data.GcpWorkloadIdentityPoolId.ValueStringPointer(),
			GcpPoolProviderId:      data.GcpPoolProviderId.ValueStringPointer(),
			ServiceAccountEmail:    data.ServiceAccountEmail.ValueStringPointer(),
		},
	}

	stores, err := r.api.SecretsHubAPI.GetGcpSecretStores(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret stores", err.Error())
		return
	}

	for _, store := range stores.SecretStores {
		if *store.Name == data.Name.ValueString() && *store.Data.GcpProjectName == data.GcpProjectName.ValueString() {
			tflog.Info(ctx, fmt.Sprintf("Secret store with name %s and Gcp Project %s already exists", data.Name.ValueString(), data.GcpProjectName.ValueString()))

			// We assume that secret store is already created
			data.ID = types.StringValue(store.ID)
			data.LastUpdated = types.StringPointerValue(store.UpdatedAt)
			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	output, err := r.api.SecretsHubAPI.AddGcpSecretStore(ctx, newStore)
	if err != nil {
		resp.Diagnostics.AddError("Error creating secret store", err.Error())
		return
	}

	data.ID = types.StringValue(output.ID)
	data.LastUpdated = types.StringPointerValue(output.UpdatedAt)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read the resource state.
func (r *gcpSecretStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data gcpSecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.api.SecretsHubAPI.GetGcpSecretStore(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret store", err.Error())
		return
	}

	data = gcpSecretStoreModel{
		Name:        types.StringPointerValue(output.Name),
		Description: types.StringPointerValue(output.Description),
		Type:        types.StringPointerValue(output.Type),
		ID:          types.StringValue(output.ID),
		LastUpdated: types.StringPointerValue(output.UpdatedAt),
	}

	if output.Data != nil {
		data.GcpProjectName = types.StringPointerValue(output.Data.GcpProjectName)
		data.GcpProjectNumber = types.StringPointerValue(output.Data.GcpProjectNumber)
		data.GcpWorkloadIdentityPoolId = types.StringPointerValue(output.Data.GcpWorkloadIdentityPoolId)
		data.GcpPoolProviderId = types.StringPointerValue(output.Data.GcpPoolProviderId)
		data.ServiceAccountEmail = types.StringPointerValue(output.Data.ServiceAccountEmail)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *gcpSecretStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state gcpSecretStoreModel

	// Read Terraform plan data and current state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prevent GCP Project Number from being updated
	if !data.GcpProjectNumber.Equal(state.GcpProjectNumber) {
		resp.Diagnostics.AddError("Invalid Update",
			"GCP Project Number cannot be changed.")
		return
	}

	updatedStore := cybrapi.SecretStoreInput[cybrapi.GcpData]{
		Name:        data.Name.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
		Data: &cybrapi.GcpData{
			GcpProjectName:  data.GcpProjectName.ValueStringPointer(),
			GcpWorkloadIdentityPoolId:  data.GcpWorkloadIdentityPoolId.ValueStringPointer(),
			GcpPoolProviderId:      data.GcpPoolProviderId.ValueStringPointer(),
			ServiceAccountEmail:    data.ServiceAccountEmail.ValueStringPointer(),
		},
	}

	output, err := r.api.SecretsHubAPI.UpdateGcpSecretStore(ctx, state.ID.ValueString(), updatedStore)
	if err != nil {
		resp.Diagnostics.AddError("Error updating secret store", err.Error())
		return
	}

	data.ID = types.StringValue(output.ID)
	data.LastUpdated = types.StringPointerValue(output.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *gcpSecretStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gcpSecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.api.SecretsHubAPI.DeleteSecretStore(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting secret store", err.Error())
		return
	}
}

func (r *gcpSecretStoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
