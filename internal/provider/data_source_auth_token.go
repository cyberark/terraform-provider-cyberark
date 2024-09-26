// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &tokenDataSource{}
	_ datasource.DataSourceWithConfigure = &tokenDataSource{}
)

// NewTokenDataSource is a helper function to simplify the provider implementation.
func NewTokenDataSource() datasource.DataSource {
	return &tokenDataSource{}
}

// tokenDataSource is the resource implementation.
type tokenDataSource struct {
	client *cybrapi.Client
}

type tokenDataSourceModel struct {
	Token types.String `tfsdk:"token"`
}

// Metadata returns the resource type name.
func (d *tokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_token"
}

// Schema returns the resource schema.
func (d *tokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Shared Services Auth Token",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "Shared Services Authorization Token",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure adds the provider configured client to this datasource.
func (d *tokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cybrapi.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *cybrapi.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Return token
func (d *tokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tokenDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Token = types.StringValue(string(d.client.AuthToken))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
