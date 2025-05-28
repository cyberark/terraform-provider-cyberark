// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"
	"time"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &secretStoreStateResource{}
	_ resource.ResourceWithConfigure      = &secretStoreStateResource{}
	_ resource.ResourceWithValidateConfig = &secretStoreStateResource{}
	_ resource.ResourceWithImportState    = &secretStoreStateResource{}
)

// NewSecretStoreStateResource is a helper function to simplify the provider implementation.
func NewSecretStoreStateResource() resource.Resource {
	return &secretStoreStateResource{}
}

// secretStoreStateResource defines the resource implementation.
type secretStoreStateResource struct {
	api *cybrapi.API
}

// secretStoreStateModel describes the resource data model.
type secretStoreStateModel struct {
	StoreID     types.String `tfsdk:"store_id"`
	Action      types.String `tfsdk:"action"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *secretStoreStateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_store_state"
}

// Schema returns the resource schema.
func (r *secretStoreStateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Secret Store State Resource

This resource is responsible for enabling or disabling a Secret Store in CyberArk Secrets Hub.

For more information click [here](https://api-docs.cyberark.com/docs/secretshub-api/qb5o0s8br9nxg-set-secret-store-state).`,
		Attributes: map[string]schema.Attribute{
			"store_id": schema.StringAttribute{
				Description: "ID of an existing CyberArk Secrets Hub Secret Store.",
				Required:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"action": schema.StringAttribute{
				Description: "Valid values are `enable` or `disable`. This field is required.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *secretStoreStateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*cybrapi.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected error configuring provider",
			fmt.Sprintf("Expected *cybrapi.Api, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.api = api
}

// ValidateConfig validates the resource configuration.
func (r *secretStoreStateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data secretStoreStateModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the action
	switch data.Action.ValueString() {
	case "enable": // valid
	case "disable": // valid
	default:
		resp.Diagnostics.AddError("Invalid Action",
			fmt.Sprintf("Action must be either 'enable' or 'disable', got: %s", data.Action.ValueString()))
	}
}

// Create a new resource.
func (r *secretStoreStateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data secretStoreStateModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.api.SecretsHubAPI.SetSecretStoreState(ctx, data.StoreID.ValueString(), data.Action.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error setting secret store state",
			fmt.Sprintf("Error setting secret store state: %+v", err))
		return
	}

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read the resource state. The Read method is used to sync an existing resource with Terraform's state when Terraform is already aware of the resource.
func (r *secretStoreStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Error(ctx, "Read method is not available for secret store state resource")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *secretStoreStateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state secretStoreStateModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.api.SecretsHubAPI.SetSecretStoreState(ctx, data.StoreID.ValueString(), data.Action.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating secret store state",
			fmt.Sprintf("Error updating secret store state: %+v", err))
		return
	}

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretStoreStateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Error(ctx, "Delete method is not available for secret store state resource")
}

func (r *secretStoreStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Error(ctx, "Import method is not available for secret store state resource")
}
