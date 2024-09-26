// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &awsSecretStoreResource{}
	_ resource.ResourceWithConfigure = &awsSecretStoreResource{}
)

// NewAWSSecretStoreResource is a helper function to simplify the provider implementation.
func NewAWSSecretStoreResource() resource.Resource {
	return &awsSecretStoreResource{}
}

// awsSecretStoreResource defines the resource implementation.
type awsSecretStoreResource struct {
	api *cybrapi.API
}

// awsSecretStoreModel describes the resource data model.
type awsSecretStoreModel struct {
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	AccountAlias types.String `tfsdk:"aws_account_alias"`
	AccountID    types.String `tfsdk:"aws_account_id"`
	RegionID     types.String `tfsdk:"aws_account_region"`
	RoleName     types.String `tfsdk:"aws_iam_role"`
	ID           types.String `tfsdk:"id"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *awsSecretStoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_secret_store"
}

// Schema returns the resource schema.
func (r *awsSecretStoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `AWS Secret Store Resource

This resource is responsible for creating a new AWS secret store in Cyberark SecretsHub.

For more information click [here](https://docs.cyberark.com/secrets-hub-privilege-cloud/Latest/en/Content/Developer/sh-create-aws-target-tutorial.htm?tocpath=Developer%7CTutorials%7C_____1).`,
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
				Description: "Should always be 'AWS_ASM' for AWS Secret Manager.",
				Computed:    true,
				Default:     stringdefault.StaticString("AWS_ASM"),
			},
			"aws_account_alias": schema.StringAttribute{
				Description: "AWS Account Alias ",
				Required:    true,
			},
			"aws_account_id": schema.StringAttribute{
				Description: "AWS Account ID",
				Required:    true,
			},
			"aws_account_region": schema.StringAttribute{
				Description: "AWS Region ID",
				Required:    true,
			},
			"aws_iam_role": schema.StringAttribute{
				Description: "AWS Role Name",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *awsSecretStoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *awsSecretStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data awsSecretStoreModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	description := data.Description.ValueString()
	storeType := data.Type.ValueString()
	accountAlias := data.AccountAlias.ValueString()
	accountID := data.AccountID.ValueString()
	regionID := data.RegionID.ValueString()
	roleName := data.RoleName.ValueString()

	newAccount := cybrapi.SecretStoreInput[cybrapi.AwsAsmData]{
		Name:        &name,
		Description: &description,
		Type:        &storeType,
		Data: &cybrapi.AwsAsmData{
			AccountAlias: &accountAlias,
			AccountID:    &accountID,
			RegionID:     &regionID,
			RoleName:     &roleName,
		},
	}

	stores, err := r.api.SecretsHubAPI.GetAwsAsmSecretStores(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret stores",
			fmt.Sprintf("Error while reading secret stores: %+v", err))
		return
	}

	for _, store := range stores.SecretStores {
		if *store.Name == name && *store.Data.AccountAlias == accountAlias {
			tflog.Info(ctx, fmt.Sprintf("Secret store with name %s and account alias %s already exists", name, accountAlias))

			// We assume that secret store is already created
			data.ID = types.StringValue(store.ID)
			data.LastUpdated = types.StringPointerValue(store.UpdatedAt)
			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	output, err := r.api.SecretsHubAPI.AddAwsAsmSecretStore(ctx, newAccount)
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
func (r *awsSecretStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data awsSecretStoreModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.api.SecretsHubAPI.GetAwsAsmSecretStore(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret store",
			fmt.Sprintf("Error while reading secret store: %+v", err))
		return
	}

	data.AccountAlias = types.StringPointerValue(output.Data.AccountAlias)
	data.RoleName = types.StringPointerValue(output.Data.RoleName)
	data.Description = types.StringPointerValue(output.Description)
	data.Name = types.StringPointerValue(output.Name)
	data.LastUpdated = types.StringPointerValue(output.UpdatedAt)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *awsSecretStoreResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *awsSecretStoreResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete is not supported through terraform",
		"Please consult with your CyberArk Administrator to process account property updates.")
}
