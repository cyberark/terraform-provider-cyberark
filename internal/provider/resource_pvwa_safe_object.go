// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
				Computed:    true,
				Default:     stringdefault.StaticString("Managed by Terraform"),
			},
			"safe_loc": schema.StringAttribute{
				Description: "The location of the Safe in the Vault.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("\\"),
			},
			"cpm_name": schema.StringAttribute{
				Description: "The name of the CPM user who will manage the new Safe.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("PasswordManager"),
			},
			"retention": schema.Int64Attribute{
				Description: "The number of days that password versions are saved in the Safe.",
				Optional:    true,
				//Computed is needed bc the PAM API defaults to 7 if not set and breaks TF state which expects null
				Computed: true,
			},
			"retention_versions": schema.Int64Attribute{
				Description: "The number of retained versions of every password that is stored in the Safe.",
				Optional:    true,
			},
			"purge": schema.BoolAttribute{
				Description: "Whether or not to automatically purge files after the end of the Object History Retention Period defined in the Safe properties.",
				Optional:    true,
				Computed:    true,
			},
			"enable_olac": schema.BoolAttribute{
				Description: "Whether or not to enable Object Level Access Control (OLAC) for the Safe.",
				Optional:    true,
				Computed:    true,
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

func (r *pvwaSafeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data pvwaSafeResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateSafeModel(data)...)
}

// Create a new resource.
func (r *pvwaSafeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data pvwaSafeResourceModel
	var raw pvwaSafeResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &raw)...)

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Terraform resolves Computed nil int values in the raw config to 0 in the plan,
	// to prevent that, set nils explicitly when needed.
	if raw.RetentionDays.IsNull() {
		data.RetentionDays = types.Int64Null()
	}

	safeName := data.Name.ValueString()

	newSafe := mapToSafeData(data)

	tflog.Info(ctx, "Creating new safe", map[string]interface{}{"safe_name": safeName})
	safe, err := r.api.PVWAAPI.AddSafe(ctx, newSafe)

	if errMsg := handleCreateError(safeName, err); errMsg != "" {
		resp.Diagnostics.AddError("Error creating safe", errMsg)
		return
	}

	tflog.Info(ctx, "Adding member to safe", map[string]interface{}{
		"safe_name": safeName,
		"member":    data.SeedMember.ValueString(),
	})

	// TODO: Shall we revert the safe creation if adding the member fails?
	_, err = r.api.PVWAAPI.AddSafeMember(ctx, newSafe)
	if err != nil {
		resp.Diagnostics.AddError("Error adding a safe member", err.Error())
		return
	}

	data = mapToSafeResourceModel(data, safe)

	data.LastUpdated = getLastUpdated(safe.LastModificationTime)

	tflog.Info(ctx, "Successfully created safe ", map[string]interface{}{
		"safe_name": safeName,
		"safe_id":   data.ID.ValueString(),
	})

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
		resp.Diagnostics.AddError("Error reading safe", err.Error())
		return
	}

	data = mapToSafeResourceModel(data, safe)

	// Set last updated time to last refreshed time
	data.LastUpdated = getLastUpdated(safe.LastModificationTime)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *pvwaSafeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var init, data, state pvwaSafeResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &init)...)
	// Read Terraform plan data and current state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// check the value of EnableOLAC - if the read one is enabled and the plan one is disabled, block the update
	if !state.EnableOLAC.IsNull() && state.EnableOLAC.ValueBool() && !data.EnableOLAC.IsNull() && !data.EnableOLAC.ValueBool() {
		resp.Diagnostics.AddError("Error updating safe", "Disabling Object Level Access Control (OLAC) is not permitted after the Safe is added.")
		return
	}
	if resp.Diagnostics.HasError() {
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

	// Unset mutually exclusive retention fields before update
	// e.g.customer update with retention set for a safe that already has retention_versions set
	if init.RetentionVersions.ValueInt64Pointer() != nil {
		updatedSafe.RetentionDays = nil
	}
	if init.RetentionDays.ValueInt64Pointer() != nil {
		updatedSafe.RetentionVersions = nil
	}

	// Call API to update the safe
	safe, err := r.api.PVWAAPI.UpdateSafe(ctx, state.ID.ValueString(), updatedSafe)
	if errMsg := handleCreateError(data.Name.ValueString(), err); errMsg != "" {
		resp.Diagnostics.AddError("Error updating safe", errMsg)
		return
	}

	if !state.SeedMember.IsNull() && !state.SeedMType.IsNull() && !state.PermType.IsNull() {
		// Validate permission level
		if err := validatePermissionLevel(data.PermType.ValueString()); err != nil {
			resp.Diagnostics.AddError("Permission Level Error", err.Error())
			return
		}

		_, err = r.api.PVWAAPI.UpdateSafeMember(ctx, updatedSafe)
		if err != nil {
			resp.Diagnostics.AddError("Error updating safe member", err.Error())
			return
		}
	} else {
		resp.Diagnostics.AddWarning("Warning updating safe member", "safe member not found in state, skipping update")
	}

	data = mapToSafeResourceModel(data, safe)

	// Update last updated time
	data.LastUpdated = getLastUpdated(safe.LastModificationTime)

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

	// First delete the safe member if possible
	if !data.SeedMember.IsNull() && !data.SeedMType.IsNull() && !data.PermType.IsNull() {
		err := r.api.PVWAAPI.DeleteSafeMember(ctx, data.Name.ValueString(), data.SeedMember.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error deleting safe member",
				fmt.Sprintf("Error while deleting safe member: %+v", err))
			// Continue with safe deletion even if member deletion fails
		}
	} else {
		resp.Diagnostics.AddWarning("Warning deleting safe member",
			"Safe member not found in state, skipping deletion")
	}

	err := r.api.PVWAAPI.DeleteSafe(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting safe", err.Error())
		return
	}
}

func (r *pvwaSafeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapToSafeResourceModel(data pvwaSafeResourceModel, safe *cybrapi.SafeData) pvwaSafeResourceModel {
	return pvwaSafeResourceModel{
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
		EnableOLAC:        types.BoolPointerValue(safe.EnableOLAC),
	}
}

func mapToSafeData(data pvwaSafeResourceModel) cybrapi.SafeData {
	return cybrapi.SafeData{
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
}

func validateSafeModel(m pvwaSafeResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := validatePermissionLevel(m.PermType.ValueString()); err != nil {
		diags.AddError("Permission Level Error", err.Error())
	}
	if err := validateRetentionFields(m.RetentionDays.ValueInt64Pointer(), m.RetentionVersions.ValueInt64Pointer()); err != nil {
		diags.AddError("Retention Fields Error", err.Error())
	}

	return diags
}

// Check if the permission level is valid.
func validatePermissionLevel(perm string) error {
	validPerms := map[string]struct{}{
		"full":     {},
		"read":     {},
		"approver": {},
		"manager":  {},
	}
	if _, ok := validPerms[perm]; !ok {
		return fmt.Errorf("permission level (%s) is invalid. Acceptable values: full, read, approver, manager", perm)
	}
	return nil
}

// Validate that only one of retention or retention_versions is set.
func validateRetentionFields(retention, retentionVersions *int64) error {
	if retention != nil && retentionVersions != nil {
		return fmt.Errorf("only one of 'retention' or 'retention_versions' may be set. Got retention=%d, retention_versions=%d", *retention, *retentionVersions)
	}
	return nil
}

func getLastUpdated(lastModificationTime *int64) types.String {
	if lastModificationTime != nil {
		newTime := time.UnixMicro(*lastModificationTime)
		return types.StringValue(newTime.Format(time.RFC3339))
	}
	return types.StringValue(time.Now().Format(time.RFC3339))
}

func handleCreateError(safeName string, err error) string {
	if err == nil {
		return ""
	}

	// Handling safe exists error
	if strings.Contains(err.Error(), "SFWS0002") {
		return fmt.Sprintf("Safe '%s' already exists. Use import to manage existing safes", safeName)
	}

	return fmt.Sprintf("Failed to create safe '%s': %s", safeName, err.Error())
}
