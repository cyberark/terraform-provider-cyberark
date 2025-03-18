// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"
	"time"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &pvwaDBAccountResource{}
	_ resource.ResourceWithConfigure   = &pvwaDBAccountResource{}
	_ resource.ResourceWithImportState = &pvwaDBAccountResource{}
)

// NewPVWADBAccountResource is a helper function to simplify the provider implementation.
func NewPVWADBAccountResource() resource.Resource {
	return &pvwaDBAccountResource{}
}

// pvwaDBAccountResource is the resource implementation.
type pvwaDBAccountResource struct {
	api *cybrapi.API
}

// Metadata returns the resource type name.
func (r *pvwaDBAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pvwa_db_account"
}

// Schema returns the resource schema.
func (r *pvwaDBAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Database Account Resource

This resource is responsible for creating a new privileged account that contains all the required DB information as mentioned below in Privilege Access Manager.

For more information click [here](https://docs.cyberark.com/pam-self-hosted/latest/en/Content/WebServices/Add%20Account%20v10.htm).`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "CyberArk Privilege Access Manager Credential ID- Generated from CyberArk after onboarding account into a safe.",
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
				Description: "Should always be 'password' for Database Credential.",
				Computed:    true,
				Default:     stringdefault.StaticString("password"),
			},
			"secret": schema.StringAttribute{
				Description: "Password of the credential object.",
				Required:    true,
				Sensitive:   true,
			},
			"secret_name_in_secret_store": schema.StringAttribute{
				Description: "Name of the credential object.",
				Optional:    true,
			},
			"sm_manage": schema.BoolAttribute{
				Description: "Automatic Management of a credential. Optional Value.",
				Optional:    true,
			},
			"sm_manage_reason": schema.StringAttribute{
				Description: "If sm_manage is false, provide reason why credential is not managed.",
				Optional:    true,
			},
			"db_port": schema.StringAttribute{
				Description: "Database connection port.",
				Optional:    true,
			},
			"dbname": schema.StringAttribute{
				Description: "Database name.",
				Optional:    true,
			},
			"db_dsn": schema.StringAttribute{
				Description: "Database data source name.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *pvwaDBAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	api, ok := req.ProviderData.(*cybrapi.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *cybrapi.Api, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.api = api
}

// Create a new resource.
func (r *pvwaDBAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data dbCredModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newAccount := cybrapi.Credential{
		Name:       data.Name.ValueStringPointer(),
		Address:    data.Address.ValueStringPointer(),
		UserName:   data.Username.ValueStringPointer(),
		Platform:   data.Platform.ValueStringPointer(),
		SafeName:   data.Safe.ValueStringPointer(),
		SecretType: data.SecretType.ValueStringPointer(),
		Secret:     data.Secret.ValueStringPointer(),
		Props: &cybrapi.AccountProps{
			Port:                    data.DBPort.ValueStringPointer(),
			DBName:                  data.DBName.ValueStringPointer(),
			DSN:                     data.DBDSN.ValueStringPointer(),
			SecretNameInSecretStore: data.SecretNameInSecretStore.ValueStringPointer(),
		},
		SecretMgmt: &cybrapi.SecretManagement{
			AutomaticManagement:    data.Manage.ValueBoolPointer(),
			ManualManagementReason: data.ManageReason.ValueStringPointer(),
		},
	}

	accountSearch, err := r.api.PVWAAPI.FilterAccounts(
		ctx,
		"",
		[]string{
			fmt.Sprintf("safeName eq %s", data.Safe.ValueString()),
		})
	if err != nil {
		resp.Diagnostics.AddError("Error searching for account", fmt.Sprintf("Error searching for account: %+v", err))
		return
	}

	var account *cybrapi.CredentialResponse

	for _, acc := range accountSearch.Accounts {
		if *acc.Name == data.Name.ValueString() {
			account = acc
			break
		}
	}

	if account == nil {
		tflog.Info(ctx, "Account not found, creating new")
		account, err = r.api.PVWAAPI.AddAccount(ctx, newAccount)
		if err != nil {
			resp.Diagnostics.AddError("Error creating account", fmt.Sprintf("Error creating account: %+v", err))
			return
		}
	} else {
		resp.Diagnostics.AddError("Error creating account", "Account already exists")
		return
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
}

// Read the resource and sets the Terraform state.
func (r *pvwaDBAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dbCredModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.api.PVWAAPI.GetAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading account", fmt.Sprintf("Error reading account: (%+v)", err))
		return
	}

	data = dbCredModel{
		Name:                    types.StringPointerValue(newState.Name),
		Address:                 types.StringPointerValue(newState.Address),
		Username:                types.StringPointerValue(newState.UserName),
		Platform:                types.StringPointerValue(newState.Platform),
		Safe:                    types.StringPointerValue(newState.SafeName),
		SecretType:              types.StringPointerValue(newState.SecretType),
		Secret:                  types.StringPointerValue(newState.Secret),
		ID:                      types.StringPointerValue(newState.CredID),
		DBPort:                  types.StringPointerValue(newState.Props.Port),
		DBName:                  types.StringPointerValue(newState.Props.DBName),
		DBDSN:                   types.StringPointerValue(newState.Props.DSN),
		SecretNameInSecretStore: types.StringPointerValue(newState.Props.SecretNameInSecretStore),
		Manage:                  types.BoolPointerValue(newState.SecretMgmt.AutomaticManagement),
		ManageReason:            types.StringPointerValue(newState.SecretMgmt.ManualManagementReason),
	}

	// Set last updated time to last updated time in the vault
	if newState.SecretMgmt != nil && newState.SecretMgmt.ModifiedTime != nil {
		newTime := time.UnixMicro(*newState.SecretMgmt.ModifiedTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *pvwaDBAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state dbCredModel

	// Read Terraform plan data and current state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updatedAccount := cybrapi.Credential{
		Name:       data.Name.ValueStringPointer(),
		Address:    data.Address.ValueStringPointer(),
		UserName:   data.Username.ValueStringPointer(),
		Platform:   data.Platform.ValueStringPointer(),
		SafeName:   data.Safe.ValueStringPointer(),
		SecretType: data.SecretType.ValueStringPointer(),
		Secret:     data.Secret.ValueStringPointer(),
		Props: &cybrapi.AccountProps{
			Port:                    data.DBPort.ValueStringPointer(),
			DBName:                  data.DBName.ValueStringPointer(),
			DSN:                     data.DBDSN.ValueStringPointer(),
			SecretNameInSecretStore: data.SecretNameInSecretStore.ValueStringPointer(),
		},
		SecretMgmt: &cybrapi.SecretManagement{
			AutomaticManagement:    data.Manage.ValueBoolPointer(),
			ManualManagementReason: data.ManageReason.ValueStringPointer(),
		},
	}

	account, err := r.api.PVWAAPI.UpdateAccount(ctx, state.ID.ValueString(), updatedAccount)
	if err != nil {
		resp.Diagnostics.AddError("Error updating account",
			fmt.Sprintf("Error while updating account: %+v", err))
		return
	}

	// Update the ID in case it changed
	data.ID = types.StringPointerValue(account.CredID)

	// Update last updated time
	if account.SecretMgmt != nil && account.SecretMgmt.ModifiedTime != nil {
		newTime := time.UnixMicro(*account.SecretMgmt.ModifiedTime)
		data.LastUpdated = types.StringValue(newTime.Format(time.RFC3339))
	} else {
		data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	}

	tflog.Info(ctx, "Database Account updated successfully")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *pvwaDBAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data dbCredModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.api.PVWAAPI.DeleteAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting account",
			fmt.Sprintf("Error while deleting account: %+v", err))
		return
	}

	tflog.Info(ctx, "Database Account deleted successfully")
}

// ImportState imports an existing resource into Terraform.
func (r *pvwaDBAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
