// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &safeResource{}
	_ resource.ResourceWithConfigure = &safeResource{}
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

This resource is responsible for creating a new privileged cloud safe in CyberaArk Privilege Cloud.

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
			"Unexpected CreateAzureAkvData Source Configure Type",
			fmt.Sprintf("Expected *cybrapi.Api, got: %T. Please report this issue to the provider developers", req.ProviderData),
		)
		return
	}

	r.api = api
}

// Create a new resource.
func (r *safeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data safeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	safeName := data.Name.ValueString()
	member := data.SeedMember.ValueString()
	memberType := data.SeedMType.ValueString()
	permissionLevel := data.PermType.ValueString()
	switch permissionLevel {
	case "full", "read", "approver", "manager":
		// valid options
	default:
		resp.Diagnostics.AddError("Permission Level Error",
			fmt.Sprintf("Permission level (%s) does not match acceptable values", data.PermType.ValueString()))
		return
	}
	// Required attributes met
	newSafe := cybrapi.SafeData{
		Name:      &safeName,
		Owner:     &member,
		OwnerType: &memberType,
		Level:     &permissionLevel,
	}

	// Processing optionals
	newSafe.Description = data.Description.ValueStringPointer()
	newSafe.Location = data.Location.ValueStringPointer()
	newSafe.CPM = data.CPM.ValueStringPointer()
	newSafe.PurgeEnabled = data.PurgeEnabled.ValueBoolPointer()
	newSafe.RetentionDays = data.RetentionDays.ValueInt64Pointer()
	newSafe.RetentionVersions = data.RetentionVersions.ValueInt64Pointer()

	// Check if there is an existing Safe
	safe, err := r.api.PamAPI.GetSafe(ctx, safeName)
	if err != nil {
		tflog.Info(ctx, "Safe not found, creating new")
		safe, err = r.api.PamAPI.AddSafe(ctx, newSafe)
		if err != nil {
			resp.Diagnostics.AddError("Error creating Safe", fmt.Sprintf("Error onboarding new Safe: (%+v)", err))
			return
		}
	}

	err = r.api.PamAPI.AddSafeMember(ctx, newSafe)
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
		resp.Diagnostics.AddError("Error reading Safe", fmt.Sprintf("Error reading Safe from API: (%+v)", err))
		return
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
func (r *safeResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *safeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data safeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the API to delete the safe
	err := r.api.PamAPI.DeleteSafe(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Safe", fmt.Sprintf("Error deleting Safe: (%+v)", err))
		return
	}

	// Remove the resource from the state
	resp.State.RemoveResource(ctx)
}
