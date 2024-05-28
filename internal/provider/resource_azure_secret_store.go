// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"

	cybrapi "github.com/cyberark/terraform-provider-secretshub/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &azureSecretStoreResource{}
	_ resource.ResourceWithConfigure = &azureSecretStoreResource{}
)

// NewAzureSecretStoreResource is a helper function to simplify the provider implementation.
func NewAzureSecretStoreResource() resource.Resource {
	return &azureSecretStoreResource{}
}

// azureSecretStoreResource defines the resource implementation.
type azureSecretStoreResource struct {
	api *cybrapi.API
}

// azureSecretStoreModel describes the resource data model.
type azureSecretStoreModel struct {
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Type                 types.String `tfsdk:"type"`
	AppClientDirectoryID types.String `tfsdk:"azure_app_client_directory_id"`
	AzureVaultURL        types.String `tfsdk:"azure_vault_url"`
	AppClientID          types.String `tfsdk:"azure_app_client_id"`
	AppClientSecret      types.String `tfsdk:"azure_app_client_secret"`
	ConnectionType       types.String `tfsdk:"connection_type"`
	ConnectorID          types.String `tfsdk:"connector_id"`
	SubscriptionID       types.String `tfsdk:"subscription_id"`
	SubscriptionName     types.String `tfsdk:"subscription_name"`
	ResourceGroupName    types.String `tfsdk:"resource_group_name"`
	ID                   types.String `tfsdk:"id"`
	LastUpdated          types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *azureSecretStoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_secret_store"
}

// Schema returns the resource schema.
func (r *azureSecretStoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Microsoft Azure Secret Store Resource

This resource is responsible for creating a new Azure secret store in Cyberark SecretsHub.

For more information click [here](https://docs.cyberark.com/secrets-hub-privilege-cloud/Latest/en/Content/Developer/sh-create-azure-store.htm?tocpath=Developer%7CTutorials%7CCreate%20an%20Azure%20secret%20store%20-%20tutorial%7C_____0).`,
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
				Description: "Should always be 'AZURE_AKV' for Azure Key Vault.",
				Computed:    true,
				Default:     stringdefault.StaticString("AZURE_AKV"),
			},
			"azure_app_client_directory_id": schema.StringAttribute{
				Description: "Azure Application Directory ID ",
				Required:    true,
			},
			"azure_vault_url": schema.StringAttribute{
				Description: "Azure Vault URL.",
				Required:    true,
			},
			"azure_app_client_id": schema.StringAttribute{
				Description: "Azure APP client ID.",
				Required:    true,
				// Sensitive:   true,
			},
			"azure_app_client_secret": schema.StringAttribute{
				Description: "Azure App Client Secret.",
				Required:    true,
				Sensitive:   true,
			},
			"connection_type": schema.StringAttribute{
				Description: "Azure Connector Type.",
				Required:    true,
			},
			"connector_id": schema.StringAttribute{
				Description: "Azure ConnectorID.",
				Required:    true,
			},
			"subscription_id": schema.StringAttribute{
				Description: "Azure SubscriptionID.",
				Required:    true,
			},
			"subscription_name": schema.StringAttribute{
				Description: "Azure Subscription Name.",
				Required:    true,
			},
			"resource_group_name": schema.StringAttribute{
				Description: "Azure resource Group Name.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *azureSecretStoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *azureSecretStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data azureSecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	description := data.Description.ValueString()
	storeType := data.Type.ValueString()
	appClientDirectoryID := data.AppClientDirectoryID.ValueString()
	azureVaultURL := data.AzureVaultURL.ValueString()
	appClientID := data.AppClientID.ValueString()
	appClientSecret := data.AppClientSecret.ValueString()

	connectionType := data.ConnectionType.ValueString()
	connectorID := data.ConnectorID.ValueString()
	subscriptionID := data.SubscriptionID.ValueString()
	subscriptionName := data.SubscriptionName.ValueString()
	resourceGroupName := data.ResourceGroupName.ValueString()
	newAccount := cybrapi.SecretStoreInput[cybrapi.CreateAzureAkvData]{
		Name:        &name,
		Description: &description,
		Type:        &storeType,
		Data: &cybrapi.CreateAzureAkvData{
			AppClientDirectoryID: &appClientDirectoryID,
			AzureVaultURL:        &azureVaultURL,
			AppClientID:          &appClientID,
			AppClientSecret:      &appClientSecret,
			Connector: &cybrapi.Connector{
				ConnectionType: &connectionType,
				ConnectorID:    &connectorID,
			},
			SubscriptionID:    &subscriptionID,
			SubscriptionName:  &subscriptionName,
			ResourceGroupName: &resourceGroupName,
		},
	}

	stores, err := r.api.SecretsHubAPI.GetAzureAkvSecretStores(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret stores",
			fmt.Sprintf("Error while reading secret stores: %+v", err))
		return
	}

	for _, store := range stores.SecretStores {
		if *store.Name == name && *store.Data.AppClientID == appClientID {
			tflog.Info(ctx, fmt.Sprintf("Secret store with name %s and account ID %s already exists", name, appClientID))

			// We assume that secret store is already created
			data.ID = types.StringValue(store.ID)
			data.LastUpdated = types.StringPointerValue(store.UpdatedAt)
			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	output, err := r.api.SecretsHubAPI.AddAzureAkvSecretStore(ctx, newAccount)
	if err != nil {
		resp.Diagnostics.AddError("Error creating secret store",
			fmt.Sprintf("Error while creating secret store: %+v", err))
		return
	}

	scanInputBody := cybrapi.TriggerScanInputBody{
		Scope: cybrapi.ScanScope{
			Scan: []string{output.ID},
		},
	}

	_, err = r.api.SecretsHubAPI.ScanDefinition(ctx, scanInputBody)
	if err != nil {
		resp.Diagnostics.AddError("Error triggering scan",
			fmt.Sprintf("Error while triggering scan: %+v", err))
		return
	}

	tflog.Info(ctx, "Secret Store created successfully")

	data.ID = types.StringValue(output.ID)
	data.LastUpdated = types.StringPointerValue(output.UpdatedAt)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read the resource state.
func (r *azureSecretStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data azureSecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.api.SecretsHubAPI.GetAzureAkvSecretStore(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret store",
			fmt.Sprintf("Error while reading secret store: %+v", err))
		return
	}

	data.AppClientDirectoryID = types.StringPointerValue(output.Data.AppClientDirectoryID)
	data.AppClientID = types.StringPointerValue(output.Data.AppClientID)
	data.AppClientSecret = types.StringPointerValue(output.Data.AppClientSecret)
	data.SubscriptionID = types.StringPointerValue(output.Data.SubscriptionID)
	data.SubscriptionName = types.StringPointerValue(output.Data.SubscriptionName)
	data.ResourceGroupName = types.StringPointerValue(output.Data.ResourceGroupName)
	data.LastUpdated = types.StringPointerValue(output.UpdatedAt)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *azureSecretStoreResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *azureSecretStoreResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}
