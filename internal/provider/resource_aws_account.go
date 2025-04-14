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
	_ resource.Resource                = &awsAccountResource{}
	_ resource.ResourceWithConfigure   = &awsAccountResource{}
	_ resource.ResourceWithImportState = &awsAccountResource{}
)

// NewAWSAccountResource is a helper function to simplify the provider implementation.
func NewAWSAccountResource() resource.Resource {
	return &awsAccountResource{}
}

// awsAccountResource defines the resource implementation.
type awsAccountResource struct {
	api *cybrapi.API
}

// awsCredModel describes the resource data model.
type awsCredModel struct {
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
	AWSKID                  types.String `tfsdk:"aws_kid"`
	AWSAccount              types.String `tfsdk:"aws_account_id"`
	Alias                   types.String `tfsdk:"aws_alias"`
	Region                  types.String `tfsdk:"aws_account_region"`
	SecretNameInSecretStore types.String `tfsdk:"secret_name_in_secret_store"`
}

// Metadata returns the resource type name.
func (r *awsAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_account"
}

// Schema returns the resource schema.
func (r *awsAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `AWS Account Resource

This resource is responsible for creating a new privileged account that contains all the required AWS information as mentioned below in Privilege Cloud.

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
				Optional:    true,
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
				Description: "Should always be 'key' for AWS Accounts.",
				Computed:    true,
				// for AWS Accounts this value must be set to key
				Default: stringdefault.StaticString("key"),
			},
			"secret": schema.StringAttribute{
				Description: "Secret Key of the credential object.",
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
			"aws_kid": schema.StringAttribute{
				Description: "AWS Access Key ID.",
				Required:    true,
			},
			"aws_account_id": schema.StringAttribute{
				Description: "AWS Account ID Number.",
				Required:    true,
			},
			"aws_alias": schema.StringAttribute{
				Description: "AWS Account Alias.",
				Optional:    true,
			},
			"aws_account_region": schema.StringAttribute{
				Description: "AWS Region.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *awsAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *awsAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data awsCredModel

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
			AWSKID:                  data.AWSKID.ValueStringPointer(),
			AWSAccount:              data.AWSAccount.ValueStringPointer(),
			Alias:                   data.Alias.ValueStringPointer(),
			Region:                  data.Region.ValueStringPointer(),
			SecretNameInSecretStore: data.SecretNameInSecretStore.ValueStringPointer(),
		},
		SecretMgmt: &cybrapi.SecretManagement{
			AutomaticManagement:    data.Manage.ValueBoolPointer(),
			ManualManagementReason: data.ManageReason.ValueStringPointer(),
		},
	}

	accountSearch, err := r.api.PamAPI.FilterAccounts(
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
		account, err = r.api.PamAPI.AddAccount(ctx, newAccount)
		if err != nil {
			resp.Diagnostics.AddError("Error creating account", fmt.Sprintf("Error creating account: %+v", err))
			return
		}
	} else {
		resp.Diagnostics.AddError("Error creating account", "Account already exist")
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

// Refresh Existing State
func (r *awsAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data awsCredModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.api.PamAPI.GetAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading account", fmt.Sprintf("Error reading account from API: (%+v)", err))
		return
	}

	data = awsCredModel{
		Name:                    types.StringPointerValue(newState.Name),
		Address:                 types.StringPointerValue(newState.Address),
		Username:                types.StringPointerValue(newState.UserName),
		Platform:                types.StringPointerValue(newState.Platform),
		Safe:                    types.StringPointerValue(newState.SafeName),
		SecretType:              types.StringPointerValue(newState.SecretType),
		ID:                      types.StringPointerValue(newState.CredID),
		Secret:                  data.Secret, // Secret is not returned by the API
		Manage:                  types.BoolPointerValue(newState.SecretMgmt.AutomaticManagement),
		ManageReason:            types.StringPointerValue(newState.SecretMgmt.ManualManagementReason),
		AWSKID:                  types.StringPointerValue(newState.Props.AWSKID),
		AWSAccount:              types.StringPointerValue(newState.Props.AWSAccount),
		Alias:                   types.StringPointerValue(newState.Props.Alias),
		Region:                  types.StringPointerValue(newState.Props.Region),
		SecretNameInSecretStore: types.StringPointerValue(newState.Props.SecretNameInSecretStore),
	}

	// Set last updated time to last updated tim in the vault
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
func (r *awsAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state awsCredModel

	// Read Terraform plan data and current state into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updatedAccount := cybrapi.Credential{
		Name:     data.Name.ValueStringPointer(),
		Address:  data.Address.ValueStringPointer(),
		UserName: data.Username.ValueStringPointer(),
		Platform: data.Platform.ValueStringPointer(),
		SafeName: data.Safe.ValueStringPointer(),
		// SecretType can not be updated
		// Secret can not be updated
		Props: &cybrapi.AccountProps{
			AWSKID:     data.AWSKID.ValueStringPointer(),
			AWSAccount: data.AWSAccount.ValueStringPointer(),
			Alias:      data.Alias.ValueStringPointer(),
			// Region can not be updated
			SecretNameInSecretStore: data.SecretNameInSecretStore.ValueStringPointer(),
		},
		SecretMgmt: &cybrapi.SecretManagement{
			AutomaticManagement:    data.Manage.ValueBoolPointer(),
			ManualManagementReason: data.ManageReason.ValueStringPointer(),
		},
	}

	account, err := r.api.PamAPI.UpdateAccount(ctx, state.ID.ValueString(), updatedAccount)
	if err != nil {
		resp.Diagnostics.AddError("Error updating account",
			fmt.Sprintf("Error updating account: %+v", err))
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

	tflog.Info(ctx, "AWS Account updated successfully")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *awsAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data awsCredModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.api.PamAPI.DeleteAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting account",
			fmt.Sprintf("Error deleting account: %+v", err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("AWS Account with ID %s deleted successfully", data.ID.ValueString()))
}

func (r *awsAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
