// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"
	"time"

	cybrapi "github.com/cyberark/terraform-provider-secretshub/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &dbAccountResource{}
	_ resource.ResourceWithConfigure = &dbAccountResource{}
)

// NewDBAccountResource is a helper function to simplify the provider implementation.
func NewDBAccountResource() resource.Resource {
	return &dbAccountResource{}
}

// dbAccountResource is the resource implementation.
type dbAccountResource struct {
	api *cybrapi.API
}

// dbCredModel describes the resource data model.
type dbCredModel struct {
	Name                    types.String `tfsdk:"name"`
	Address                 types.String `tfsdk:"address"`
	Username                types.String `tfsdk:"username"`
	Platform                types.String `tfsdk:"platform"`
	Safe                    types.String `tfsdk:"safe"`
	SecretType              types.String `tfsdk:"secret_type"`
	Secret                  types.String `tfsdk:"secret"`
	ID                      types.String `tfsdk:"id"`
	LastUpdated             types.String `tfsdk:"last_updated"`
	DBPort                  types.String `tfsdk:"db_port"`
	DBName                  types.String `tfsdk:"dbname"`
	DBDSN                   types.String `tfsdk:"db_dsn"`
	SecretNameInSecretStore types.String `tfsdk:"secret_name_in_secret_store"`

	Manage       types.Bool   `tfsdk:"sm_manage"`
	ManageReason types.String `tfsdk:"sm_manage_reason"`
}

// Metadata returns the resource type name.
func (r *dbAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_db_account"
}

// Schema returns the resource schema.
func (r *dbAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Database Account Resource

This resource is responsible for creating a new privileged account that contains all the required DB information as mentioned below in Privilege Cloud.

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
				Required:    true,
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
func (r *dbAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *dbAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data dbCredModel
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
	props.Port = data.DBPort.ValueStringPointer()
	props.DBName = data.DBName.ValueStringPointer()
	props.DSN = data.DBDSN.ValueStringPointer()
	props.SecretNameInSecretStore = data.SecretNameInSecretStore.ValueStringPointer()
	smProps.AutomaticManagement = data.Manage.ValueBoolPointer()
	smProps.ManualManagementReason = data.ManageReason.ValueStringPointer()

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
}

// Read the resource and sets the Terraform state.
func (r *dbAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data dbCredModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.api.PamAPI.GetAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading account", fmt.Sprintf("Error reading account: (%+v)", err))
		return
	}

	// Main Credentials
	data.Name = types.StringPointerValue(newState.Name)

	data.Address = types.StringPointerValue(newState.Address)
	data.Platform = types.StringPointerValue(newState.Platform)
	data.Safe = types.StringPointerValue(newState.SafeName)
	data.Username = types.StringPointerValue(newState.UserName)
	data.SecretType = types.StringPointerValue(newState.SecretType)

	// DB Props
	if newState.Props != nil {
		data.DBDSN = types.StringPointerValue(newState.Props.DSN)
		data.DBPort = types.StringPointerValue(newState.Props.Port)
		data.DBName = types.StringPointerValue(newState.Props.DBName)
	}

	// Set last updated time to last updated time in the vault
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
func (r *dbAccountResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *dbAccountResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}
