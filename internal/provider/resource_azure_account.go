// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"
	"time"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &azureAccountResource{}
	_ resource.ResourceWithConfigure = &azureAccountResource{}
)

// NewAzureAccountResource is a helper function to simplify the provider implementation.
func NewAzureAccountResource() resource.Resource {
	return &azureAccountResource{}
}

// azureAccountResource is the resource implementation.
type azureAccountResource struct {
	api *cybrapi.API
}

// azureCredModel describes the resource data model.
type azureCredModel struct {
	Name                    types.String `tfsdk:"name"`
	Address                 types.String `tfsdk:"address"`
	Username                types.String `tfsdk:"username"`
	Platform                types.String `tfsdk:"platform"`
	Safe                    types.String `tfsdk:"safe"`
	SecretType              types.String `tfsdk:"secret_type"`
	Secret                  types.String `tfsdk:"secret"`
	ID                      types.String `tfsdk:"id"`
	LastUpdated             types.String `tfsdk:"last_updated"`
	Manage                  types.Bool   `tfsdk:"sm_manage"`
	ManageReason            types.String `tfsdk:"sm_manage_reason"`
	MAppID                  types.String `tfsdk:"ms_app_id"`
	MAppObjectID            types.String `tfsdk:"ms_app_obj_id"`
	MKID                    types.String `tfsdk:"ms_key_id"`
	MADID                   types.String `tfsdk:"ms_ad_id"`
	MDur                    types.String `tfsdk:"ms_duration"`
	MPop                    types.String `tfsdk:"ms_pop"`
	MKeyDesc                types.String `tfsdk:"ms_key_desc"`
	SecretNameInSecretStore types.String `tfsdk:"secret_name_in_secret_store"`
}

// Metadata returns the resource type name.
func (r *azureAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_account"
}

// Schema returns the resource schema.
func (r *azureAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Microsoft Azure Account Resource

This resource is responsible for creating a new privileged account that contains all the required Azure information as mentioned below in Privilege Cloud.

For more information click [here](https://docs.cyberark.com/privilege-cloud-shared-services/latest/en/Content/WebServices/Add%20Account%20v10.htm).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "CyberArk Privilege Cloud Credential ID- Generated from CyberArk after onboarding account into a safe.",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "Custom Account Name for customizing the object name in a safe.",
				Required:    true,
			},
			"address": schema.StringAttribute{
				Description: "URI, URL or IP associated with the credential.",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username of the Credential object.",
				Required:    true,
			},
			"platform": schema.StringAttribute{
				Description: "Management Platform associated with the Database Credential.",
				Required:    true,
			},
			"safe": schema.StringAttribute{
				Description: "Target Safe where the credential object will be onboarded.",
				Required:    true,
			},
			"secret_type": schema.StringAttribute{
				Description: "Should always be 'password' for Azure Account.",
				Computed:    true,
				Default:     stringdefault.StaticString("password"),
			},
			"secret": schema.StringAttribute{
				Description: "Password of the credential object.",
				Required:    true,
				Sensitive:   true,
			},
			"sm_manage": schema.BoolAttribute{
				Description: "Automatic Management of a credential. Optional Value.",
				Optional:    true,
			},
			"secret_name_in_secret_store": schema.StringAttribute{
				Description: "Name of the credential object.",
				Optional:    true,
			},
			"sm_manage_reason": schema.StringAttribute{
				Description: "If sm_manage is false, provide reason why credential is not managed.",
				Optional:    true,
			},
			"ms_app_id": schema.StringAttribute{
				Description: "Microsoft Azure Application ID.",
				Required:    true,
			},
			"ms_app_obj_id": schema.StringAttribute{
				Description: "Microsoft Azure Application Object ID.",
				Required:    true,
			},
			"ms_key_id": schema.StringAttribute{
				Description: "Microsoft Azure Key ID.",
				Required:    true,
			},
			"ms_ad_id": schema.StringAttribute{
				Description: "Microsoft Azure Active Directory ID.",
				Optional:    true,
			},
			"ms_duration": schema.StringAttribute{
				Description: "Duration.",
				Optional:    true,
			},
			"ms_pop": schema.StringAttribute{
				Description: "Populate if not exist.",
				Optional:    true,
			},
			"ms_key_desc": schema.StringAttribute{
				Description: "Key Description.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *azureAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *azureAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data azureCredModel
	var props cybrapi.AccountProps
	var smProps cybrapi.SecretManagement

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	address := data.Address.ValueString()
	username := data.Username.ValueString()
	platform := data.Platform.ValueString()
	safe := data.Safe.ValueString()
	secretType := data.SecretType.ValueString()
	secret := data.Secret.ValueString()
	smProps.AutomaticManagement = data.Manage.ValueBoolPointer()
	smProps.ManualManagementReason = data.ManageReason.ValueStringPointer()
	props.MAppID = data.MAppID.ValueStringPointer()
	props.MAppObjectID = data.MAppObjectID.ValueStringPointer()
	props.MKID = data.MKID.ValueStringPointer()
	props.MADID = data.MADID.ValueStringPointer()
	props.MDur = data.MDur.ValueStringPointer()
	props.MPop = data.MPop.ValueStringPointer()
	props.MKeyDesc = data.MKeyDesc.ValueStringPointer()
	props.SecretNameInSecretStore = data.SecretNameInSecretStore.ValueStringPointer()

	newAccount := cybrapi.Credential{
		Name:       &name,
		Address:    &address,
		UserName:   &username,
		Platform:   &platform,
		SafeName:   &safe,
		SecretType: &secretType,
		Secret:     &secret,
		Props:      &props,
		SecretMgmt: &smProps,
	}

	accountSearch, err := r.api.PamAPI.FilterAccounts(
		ctx,
		// name,
		"",
		[]string{
			fmt.Sprintf("safeName eq %s", safe),
		})
	if err != nil {
		resp.Diagnostics.AddError("Error searching for account", fmt.Sprintf("Error searching for account: %+v", err))
		secret = ""
		return
	}

	var account *cybrapi.CredentialResponse

	for _, acc := range accountSearch.Accounts {
		if *acc.Name == name {
			account = acc
			break
		}
	}

	if account == nil {
		tflog.Info(ctx, "Account not found, creating new")
		account, err = r.api.PamAPI.AddAccount(ctx, newAccount)
		if err != nil {
			resp.Diagnostics.AddError("Error creating account", fmt.Sprintf("Error creating account: %+v", err))
			secret = ""
			return
		}
	}

	data.ID = types.StringPointerValue(account.CredID)

	// Set last updated time to last updated time in the vault
	if account.SecretMgmt != nil && account.SecretMgmt.ModifiedTime != nil {
		newTime := time.UnixMicro(*account.SecretMgmt.ModifiedTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	// Clear sensitive data
	secret = ""
}

// Read the resource and sets the Terraform state.
func (r *azureAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data azureCredModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.api.PamAPI.GetAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading account", fmt.Sprintf("Error reading account from API: %v", err))
		return
	}

	// Main Credentials
	data.Name = types.StringPointerValue(newState.Name)

	data.Address = types.StringPointerValue(newState.Address)
	data.Platform = types.StringPointerValue(newState.Platform)
	data.Safe = types.StringPointerValue(newState.SafeName)
	data.Username = types.StringPointerValue(newState.UserName)
	data.SecretType = types.StringPointerValue(newState.SecretType)

	// MS Props
	if newState.Props != nil {
		data.MAppID = types.StringPointerValue(newState.Props.MAppID)
		data.MAppObjectID = types.StringPointerValue(newState.Props.MAppObjectID)
		data.MKID = types.StringPointerValue(newState.Props.MKID)
		data.MADID = types.StringPointerValue(newState.Props.MADID)
		data.MDur = types.StringPointerValue(newState.Props.MDur)
		data.MPop = types.StringPointerValue(newState.Props.MPop)
		data.MKeyDesc = types.StringPointerValue(newState.Props.MKeyDesc)
	}

	// Set last updated time to last updated tim in the vault
	if newState.SecretMgmt != nil && newState.SecretMgmt.ModifiedTime != nil {
		newTime := time.UnixMicro(*newState.SecretMgmt.ModifiedTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	// Ensure ID is consistent
	data.ID = types.StringPointerValue(newState.CredID)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *azureAccountResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *azureAccountResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}
