// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &secretsHubProvider{}
)

const (
	cloudAuthURL       = "https://%s.id.cyberark.cloud"
	cloudPamURL        = "https://%s.privilegecloud.cyberark.cloud"
	cloudSecretsHubURL = "https://%s.secretshub.cyberark.cloud"
)

// secretsHubProvider defines the provider implementation.
type secretsHubProvider struct {
	version string
}

// secretsHubProviderModel describes the provider data model.
type secretsHubProviderModel struct {
	Tenant       types.String `tfsdk:"tenant"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Domain       types.String `tfsdk:"domain"`
}

// Metadata returns the provider type name.
func (p *secretsHubProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cyberark"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *secretsHubProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure tenant used to onboard account types into CyberArk Privilege Cloud Vault",
		Attributes: map[string]schema.Attribute{
			"tenant": schema.StringAttribute{
				Description: "CyberArk Shared Services Tenant.",
				Required:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "CyberArk Client ID, formatted as username@cyberark.cloud.tenant.",
				Required:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "CyberArk Client ID Password.",
				Required:    true,
				Sensitive:   true,
			},
			"domain": schema.StringAttribute{
				Description: "CyberArk Privilege Cloud Domain.",
				Required:    true,
			},
		},
	}
}

// Configure parses the configuration data and initializes the provider.
func (p *secretsHubProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data secretsHubProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t := data.Tenant.ValueString()
	cid := data.ClientID.ValueString()
	d := data.Domain.ValueString()

	// Create a client for Cyberark ISPSS (Identity Security Platform Shared Services)
	authAPI := cybrapi.NewAuthAPI(fmt.Sprintf(cloudAuthURL, t))

	token, err := authAPI.GetToken(ctx, cid, []byte(data.ClientSecret.ValueString()))

	if err != nil {
		resp.Diagnostics.AddError("Failed to get authentication token",
			fmt.Sprintf("Failed to get authentication token from Cyberark ISPSS service: %+v", err))
		return
	}

	// Create a client for Cyberark PAM
	pamAPI := cybrapi.NewPAMAPI(fmt.Sprintf(cloudPamURL, d), token)

	// Create a client for Cyberark SecretsHub
	secretsHubAPI := cybrapi.NewSecretsHubAPI(fmt.Sprintf(cloudSecretsHubURL, d), token)

	resp.DataSourceData = &cybrapi.API{
		PamAPI:        pamAPI,
		SecretsHubAPI: secretsHubAPI,
	}
	resp.ResourceData = &cybrapi.API{
		PamAPI:        pamAPI,
		SecretsHubAPI: secretsHubAPI,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *secretsHubProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTokenDataSource,
	}
}

// Resources define the resources implemented in the provider.
func (p *secretsHubProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAWSAccountResource,
		NewAWSSecretStoreResource,
		NewAzureAccountResource,
		NewAzureSecretStoreResource,
		NewDBAccountResource,
		NewSafeResource,
		NewSyncPolicyResource,
	}
}

// New creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &secretsHubProvider{
			version: version,
		}
	}
}
