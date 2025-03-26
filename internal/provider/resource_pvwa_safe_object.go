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
	_ resource.Resource                = &pvwaSafeResource{}
	_ resource.ResourceWithConfigure   = &pvwaSafeResource{}
	_ resource.ResourceWithImportState = &pvwaSafeResource{}
)

// NewPVWASafeResource is a helper function to simplify the provider implementation.
func NewPVWASafeResource() resource.Resource {
	return &pvwaSafeResource{}
}

// safeResource defines the resource implementation.
type pvwaSafeResource struct {
	api *cybrapi.API
}

// ExampleResourceModel describes the resource data model.
type pvwaSafeResourceModel struct {
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
	EnableOLAC        types.Bool   `tfsdk:"enable_olac"`
}

// Metadata returns the resource type name.
func (r *pvwaSafeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pvwa_safe"
}

// Schema returns the resource schema.
func (r *pvwaSafeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `CyberArk Privilege Access Manager Safe Resource

This resource is responsible for creating a new safe in CyberArk Privilege Access Manager.

For more information click [here](https://docs.cyberark.com/pam-self-hosted/latest/en/Content/WebServices/Add%20Safe.htm).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "CyberArk Privilege Access Manager Safe URL ID- Generated from CyberArk after onboarding safe.",
				Computed:    true,
			},
			"id_number": schema.Int64Attribute{
				Description: "CyberArk Privilege Access Manager Safe ID- Generated from CyberArk after onboarding safe.",
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
				Optional:    true,
			},
			"cpm_name": schema.StringAttribute{
				Description: "The name of the CPM user who will manage the new Safe.",
				Optional:    true,
			},
			"retention": schema.Int64Attribute{
				Description: "The number of retained versions of every password that is stored in the Safe.",
				Optional:    true,
			},
			"retention_versions": schema.Int64Attribute{
				Description: "The number of days that password versions are saved in the Safe.",
				Optional:    true,
			},
			"purge": schema.BoolAttribute{
				Description: "Whether or not to automatically purge files after the end of the Object History Retention Period defined in the Safe properties.",
				Optional:    true,
			},
			"enable_olac": schema.BoolAttribute{
				Description: "Whether or not to enable Object Level Access Control (OLAC) for the Safe.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *pvwaSafeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*cybrapi.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *cybrapi.Api, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		return
	}

	r.api = api
}

// Create a new resource.
func (r *pvwaSafeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data pvwaSafeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
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

	newSafe := cybrapi.SafeData{
		RetentionDays:     data.RetentionDays.ValueInt64Pointer(),
		RetentionVersions: data.RetentionVersions.ValueInt64Pointer(),
		PurgeEnabled:      data.PurgeEnabled.ValueBoolPointer(),
		CPM:               data.CPM.ValueStringPointer(),
		Name:              data.Name.ValueStringPointer(),
		Description:       data.Description.ValueStringPointer(),
		Location:          data.Location.ValueStringPointer(),
		Owner:             data.SeedMember.ValueStringPointer(),
		OwnerType:         data.SeedMType.ValueStringPointer(),
		Level:             data.PermType.ValueStringPointer(),
		EnableOLAC:        data.EnableOLAC.ValueBoolPointer(),
	}

	// Check if there is an existing Safe
	safe, err := r.api.PVWAAPI.GetSafe(ctx, data.Name.ValueString())
	if err != nil {
		tflog.Info(ctx, "Safe not found, creating new")
		safe, err = r.api.PVWAAPI.AddSafe(ctx, newSafe)
		if err != nil {
			resp.Diagnostics.AddError("Error creating Safe", fmt.Sprintf("Error onboarding new Safe: (%+v)", err))
			return
		}
	}

	err = r.api.PVWAAPI.AddSafeMember(ctx, newSafe)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Safe Member", fmt.Sprintf("Error adding member to the Safe: (%+v)", err))
		return
	}

	data.ID = types.StringPointerValue(safe.URLID)
	data.IDNUM = types.Int64PointerValue(safe.NUMBER)

	// Set last updated time to last refreshed time
	if safe.LastModificationTime != nil {
		newTime := time.UnixMicro(*safe.LastModificationTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	tflog.Info(ctx, "Safe and Safe Member created successfully")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read the resource and set the Terraform state.
func (r *pvwaSafeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data pvwaSafeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	safe, err := r.api.PVWAAPI.GetSafe(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Safe", fmt.Sprintf("Error reading Safe from API: (%+v)", err))
		return
	}

	data = pvwaSafeResourceModel{
		ID:                types.StringPointerValue(safe.URLID),
		IDNUM:             types.Int64PointerValue(safe.NUMBER),
		RetentionDays:     types.Int64PointerValue(safe.RetentionDays),
		RetentionVersions: types.Int64PointerValue(safe.RetentionVersions),
		PurgeEnabled:      types.BoolPointerValue(safe.PurgeEnabled),
		CPM:               types.StringPointerValue(safe.CPM),
		Name:              types.StringPointerValue(safe.Name),
		Description:       types.StringPointerValue(safe.Description),
		Location:          types.StringPointerValue(safe.Location),
		SeedMember:        types.StringPointerValue(safe.Owner),
		SeedMType:         types.StringPointerValue(safe.OwnerType),
		PermType:          types.StringPointerValue(safe.Level),
		EnableOLAC:        types.BoolPointerValue(safe.EnableOLAC),
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
func (r *pvwaSafeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state pvwaSafeResourceModel

	// Read Terraform plan data and current state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

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

	updatedSafe := cybrapi.SafeData{
		RetentionDays:     data.RetentionDays.ValueInt64Pointer(),
		RetentionVersions: data.RetentionVersions.ValueInt64Pointer(),
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
		EnableOLAC:        data.EnableOLAC.ValueBoolPointer(),
	}

	// Call API to update the safe
	safe, err := r.api.PVWAAPI.UpdateSafe(ctx, state.ID.ValueString(), updatedSafe)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Safe",
			fmt.Sprintf("Error while updating Safe: %+v", err))
		return
	}

	err = r.api.PVWAAPI.UpdateSafeMember(ctx, updatedSafe)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Safe Member",
			fmt.Sprintf("Error while updating Safe Member permissions: %+v", err))
		return
	}

	// Update ID fields in case they changed
	data.ID = types.StringPointerValue(safe.URLID)
	data.IDNUM = types.Int64PointerValue(safe.NUMBER)

	// Update last updated time
	if safe.LastModificationTime != nil {
		newTime := time.UnixMicro(*safe.LastModificationTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	tflog.Info(ctx, "Safe and Safe Member updated successfully")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *pvwaSafeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data pvwaSafeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First delete the safe member
	err := r.api.PVWAAPI.DeleteSafeMember(ctx, data.Name.ValueString(), data.SeedMember.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Safe Member",
			fmt.Sprintf("Error while deleting Safe Member: %+v", err))
		// Continue with safe deletion even if member deletion fails
	} else {
		tflog.Info(ctx, "Safe Member deleted successfully")
	}

	// Then delete the safe
	err = r.api.PVWAAPI.DeleteSafe(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Safe",
			fmt.Sprintf("Error while deleting Safe: %+v", err))
		return
	}

	tflog.Info(ctx, "Safe deleted successfully")
}

func (r *pvwaSafeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
