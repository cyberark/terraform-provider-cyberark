// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &safeResource{}
	_ resource.ResourceWithConfigure      = &safeResource{}
	_ resource.ResourceWithImportState    = &safeResource{}
	_ resource.ResourceWithValidateConfig = &safeResource{}
)

// NewSafeResource is a helper function to simplify the provider implementation.
func NewSafeResource() resource.Resource {
	return &safeResource{}
}

// safeResource defines the resource implementation.
type safeResource struct {
	api *cybrapi.API
}

// ExampleResourceModel describes the resource data model.
type safeResourceModel struct {
	RetentionDays     types.Int64  `tfsdk:"retention"`
	RetentionVersions types.Int64  `tfsdk:"retention_versions"`
	PurgeEnabled      types.Bool   `tfsdk:"purge"`
	CPM               types.String `tfsdk:"cpm_name"`
	Name              types.String `tfsdk:"safe_name"`
	Description       types.String `tfsdk:"safe_desc"`
	Location          types.String `tfsdk:"safe_loc"`
	ID                types.String `tfsdk:"id"`
	IDNUM             types.Int64  `tfsdk:"id_number"`
	LastUpdated       types.String `tfsdk:"last_updated"`
	SeedMember        types.String `tfsdk:"member"`
	SeedMType         types.String `tfsdk:"member_type"`
	PermType          types.String `tfsdk:"permission_level"`
}

// Metadata returns the resource type name.
func (r *safeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_safe"
}

// Schema returns the resource schema.
func (r *safeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `CyberArk Privilege Cloud Safe Resource

This resource is responsible for creating a new privileged cloud safe in CyberArk Privilege Cloud.

For more information click [here](https://docs.cyberark.com/privilege-cloud-shared-services/latest/en/Content/WebServices/Add%20Safe.htm).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "CyberArk Privilege Cloud Safe URL ID- Generated from CyberArk after onboarding safe.",
				Computed:    true,
			},
			"id_number": schema.Int64Attribute{
				Description: "CyberArk Privilege Cloud Safe ID- Generated from CyberArk after onboarding safe.",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"safe_name": schema.StringAttribute{
				Description: "The unique name of the Safe. The following characters cannot be used in the Safe name: \\ / : * < > . | ? “% & +",
				Required:    true,
			},
			"member": schema.StringAttribute{
				Description: "Owning Safe Member.",
				Required:    true,
			},
			"member_type": schema.StringAttribute{
				Description: "Member user type: user or group.",
				Required:    true,
			},
			"permission_level": schema.StringAttribute{
				Description: "Membership Permission Level. Currently supported inputs: full, read, approver, manager.",
				Required:    true,
			},
			"safe_desc": schema.StringAttribute{
				Description: "The description of the Safe.",
				Optional:    true,
			},
			"safe_loc": schema.StringAttribute{
				Description: "The location of the Safe in the Vault.",
				Computed:    true,
				Optional:    true,
			},
			"cpm_name": schema.StringAttribute{
				Description: "The name of the CPM user who will manage the new Safe.",
				Computed:    true,
				Optional:    true,
			},
			"retention": schema.Int64Attribute{
				Description: "The number of days that password versions are saved in the Safe.",
				Computed:    true,
				Optional:    true,
			},
			"retention_versions": schema.Int64Attribute{
				Description: "The number of retained versions of every password that is stored in the Safe.",
				Computed:    true,
				Optional:    true,
			},
			"purge": schema.BoolAttribute{
				Description: "Whether or not to automatically purge files after the end of the Object History Retention Period defined in the Safe properties.",
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *safeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*cybrapi.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected AzureAkvData Source Configure Type",
			fmt.Sprintf("Expected *cybrapi.Api, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		return
	}

	r.api = api
}

// ValidateConfig validates the resource configuration.
func (r *safeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data safeResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate permission level
	switch data.PermType.ValueString() {
	case "full", "read", "approver", "manager":
		// valid options
	default:
		resp.Diagnostics.AddError("Permission Level Error",
			fmt.Sprintf("Permission level (%s) does not match acceptable values", data.PermType.ValueString()))
		return
	}

	// Ensure at most one of retention or retention_versions is set
	if !data.RetentionDays.IsNull() && !data.RetentionVersions.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration", "Only one of 'retention' or 'retention_versions' may be set.")
		return
	}
}

// Create a new resource.
func (r *safeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data safeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newSafe := cybrapi.SafeData{
		RetentionDays:     nullIfUnknown(data.RetentionDays).ValueInt64Pointer(),
		RetentionVersions: nullIfUnknown(data.RetentionVersions).ValueInt64Pointer(),
		PurgeEnabled:      data.PurgeEnabled.ValueBoolPointer(),
		CPM:               data.CPM.ValueStringPointer(),
		Name:              data.Name.ValueStringPointer(),
		Description:       data.Description.ValueStringPointer(),
		Location:          data.Location.ValueStringPointer(),
		Owner:             data.SeedMember.ValueStringPointer(),
		OwnerType:         data.SeedMType.ValueStringPointer(),
		Level:             data.PermType.ValueStringPointer(),
	}

	// Check if there is an existing Safe
	safe, err := r.api.PamAPI.GetSafe(ctx, data.Name.ValueString())
	if err != nil {
		tflog.Info(ctx, "Safe not found, creating new")
		safe, err = r.api.PamAPI.AddSafe(ctx, newSafe)
		if err != nil {
			resp.Diagnostics.AddError("Error creating safe", err.Error())
			return
		}
	}

	_, err = r.api.PamAPI.AddSafeMember(ctx, newSafe)
	if err != nil {
		resp.Diagnostics.AddError("Error creating safe member", err.Error())
		return
	}

	data = safeResourceModel{
		ID:                types.StringPointerValue(safe.URLID),
		IDNUM:             types.Int64PointerValue(safe.NUMBER),
		RetentionDays:     types.Int64PointerValue(safe.RetentionDays),
		RetentionVersions: types.Int64PointerValue(safe.RetentionVersions),
		PurgeEnabled:      types.BoolPointerValue(safe.PurgeEnabled),
		CPM:               types.StringPointerValue(safe.CPM),
		Name:              types.StringPointerValue(safe.Name),
		Description:       types.StringPointerValue(safe.Description),
		Location:          types.StringPointerValue(safe.Location),
		SeedMember:        data.SeedMember, // Can not be read from API
		SeedMType:         data.SeedMType,  // Can not be read from API
		PermType:          data.PermType,   // Can not be read from API
	}

	// Set last updated time to last refreshed time
	if safe.LastModificationTime != nil {
		newTime := time.UnixMicro(*safe.LastModificationTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read the resource and set the Terraform state.
func (r *safeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data safeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	safe, err := r.api.PamAPI.GetSafe(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading safe", err.Error())
		return
	}

	data = safeResourceModel{
		ID:                types.StringPointerValue(safe.URLID),
		IDNUM:             types.Int64PointerValue(safe.NUMBER),
		RetentionDays:     types.Int64PointerValue(safe.RetentionDays),
		RetentionVersions: types.Int64PointerValue(safe.RetentionVersions),
		PurgeEnabled:      types.BoolPointerValue(safe.PurgeEnabled),
		CPM:               types.StringPointerValue(safe.CPM),
		Name:              types.StringPointerValue(safe.Name),
		Description:       types.StringPointerValue(safe.Description),
		Location:          types.StringPointerValue(safe.Location),
		SeedMember:        data.SeedMember, // Can not be read from API
		SeedMType:         data.SeedMType,  // Can not be read from API
		PermType:          data.PermType,   // Can not be read from API
	}

	// Set last updated time to last refreshed time
	if safe.LastModificationTime != nil {
		newTime := time.UnixMicro(*safe.LastModificationTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *safeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state safeResourceModel

	// Read Terraform plan data and current state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updatedSafe := cybrapi.SafeData{
		RetentionDays:     nullIfUnknown(data.RetentionDays).ValueInt64Pointer(),
		RetentionVersions: nullIfUnknown(data.RetentionVersions).ValueInt64Pointer(),
		PurgeEnabled:      data.PurgeEnabled.ValueBoolPointer(),
		CPM:               data.CPM.ValueStringPointer(),
		Name:              data.Name.ValueStringPointer(),
		Description:       data.Description.ValueStringPointer(),
		Location:          data.Location.ValueStringPointer(),
		URLID:             data.ID.ValueStringPointer(),
		NUMBER:            data.IDNUM.ValueInt64Pointer(),
		Owner:             data.SeedMember.ValueStringPointer(),
		OwnerType:         data.SeedMType.ValueStringPointer(),
		Level:             data.PermType.ValueStringPointer(),
	}

	// Call API to update the safe
	safe, err := r.api.PamAPI.UpdateSafe(ctx, state.ID.ValueString(), updatedSafe)
	if err != nil {
		resp.Diagnostics.AddError("Error updating safe", err.Error())
		return
	}

	if !data.SeedMember.IsNull() && !data.SeedMType.IsNull() && !data.PermType.IsNull() {
		// Validate permission level
		switch data.PermType.ValueString() {
		case "full", "read", "approver", "manager":
			// valid options
		default:
			resp.Diagnostics.AddError("Permission Level Error",
				fmt.Sprintf("Permission level (%s) does not match acceptable values", data.PermType.ValueString()))
			return
		}

		_, err = r.api.PamAPI.UpdateSafeMember(ctx, updatedSafe)
		if err != nil {
			resp.Diagnostics.AddError("Error updating safe member", err.Error())
			return
		}
	} else {
		resp.Diagnostics.AddWarning("Warning updating safe member", "Safe member not found in state, skipping update")
	}

	data = safeResourceModel{
		ID:                types.StringPointerValue(safe.URLID),
		IDNUM:             types.Int64PointerValue(safe.NUMBER),
		RetentionDays:     types.Int64PointerValue(safe.RetentionDays),
		RetentionVersions: types.Int64PointerValue(safe.RetentionVersions),
		PurgeEnabled:      types.BoolPointerValue(safe.PurgeEnabled),
		CPM:               types.StringPointerValue(safe.CPM),
		Name:              types.StringPointerValue(safe.Name),
		Description:       types.StringPointerValue(safe.Description),
		Location:          types.StringPointerValue(safe.Location),
		SeedMember:        data.SeedMember, // Can not be read from API
		SeedMType:         data.SeedMType,  // Can not be read from API
		PermType:          data.PermType,   // Can not be read from API
	}

	// Update last updated time
	if safe.LastModificationTime != nil {
		newTime := time.UnixMicro(*safe.LastModificationTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *safeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data safeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First delete the safe member if possible
	if !data.SeedMember.IsNull() && !data.SeedMType.IsNull() && !data.PermType.IsNull() {
		err := r.api.PamAPI.DeleteSafeMember(ctx, data.Name.ValueString(), data.SeedMember.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error deleting safe member", err.Error())
			// Continue with safe deletion even if member deletion fails
		}
	} else {
		resp.Diagnostics.AddWarning("Warning deleting safe member", "Safe member not found in state, skipping deletion")
	}

	// Then delete the safe
	err := r.api.PamAPI.DeleteSafe(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting safe", err.Error())
		return
	}
}

func (r *safeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper method to translate unknown int64 values to null (as opposed to 0)
func nullIfUnknown(v types.Int64) types.Int64 {
	if v.IsUnknown() {
		return types.Int64Null()
	}
	return v
}
