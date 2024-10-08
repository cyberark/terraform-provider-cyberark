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
	_ provider.Provider                   = &secretsHubProvider{}
	_ provider.ProviderWithValidateConfig = &secretsHubProvider{}
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
	Tenant          types.String `tfsdk:"tenant"`
	ClientID        types.String `tfsdk:"client_id"`
	ClientSecret    types.String `tfsdk:"client_secret"`
	Domain          types.String `tfsdk:"domain"`
	PVWAUsername    types.String `tfsdk:"pvwa_username"`
	PVWAPassword    types.String `tfsdk:"pvwa_password"`
	PVWAURL         types.String `tfsdk:"pvwa_url"`
	PVWALoginMethod types.String `tfsdk:"pvwa_login_method"`
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
			"pvwa_username": schema.StringAttribute{
				Description: "CyberArk PVWA Username.",
				Optional:    true,
			},
			"pvwa_password": schema.StringAttribute{
				Description: "CyberArk PVWA Password.",
				Optional:    true,
				Sensitive:   true,
			},
			"pvwa_url": schema.StringAttribute{
				Description: "CyberArk PVWA URL.",
				Optional:    true,
			},
			"pvwa_login_method": schema.StringAttribute{
				Description: "CyberArk PVWA Login Method.",
				Optional:    true,
			},
		},
	}
}

// ValidateConfig validates the configuration data.
func (p *secretsHubProvider) ValidateConfig(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var data secretsHubProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate PVWA Login Method
	validPVWALoginMethods := []string{"cyberark", "ldap", "windows", "radius"}
	if data.PVWALoginMethod.ValueString() != "" {
		valid := false
		for _, method := range validPVWALoginMethods {
			if data.PVWALoginMethod.ValueString() == method {
				valid = true
				break
			}
		}

		if !valid {
			resp.Diagnostics.AddError("Invalid PVWA Login Method",
				fmt.Sprintf("Invalid PVWA Login Method: %s. Valid methods are: %v", data.PVWALoginMethod.ValueString(), validPVWALoginMethods))
		}
	}

	// Validate PVWA attributes (not including PVWA Login Method which defaults to "cyberark")
	pvwaAttributes := map[string]types.String{
		"pvwa_username": data.PVWAUsername,
		"pvwa_password": data.PVWAPassword,
		"pvwa_url":      data.PVWAURL,
	}

	// Check if any PVWA attribute is set
	anySet := false
	for _, attr := range pvwaAttributes {
		if attr.ValueString() != "" {
			anySet = true
			break
		}
	}

	// If any PVWA attribute is set, ensure all are set
	if anySet {
		for name, attr := range pvwaAttributes {
			if attr.ValueString() == "" {
				resp.Diagnostics.AddError("Missing PVWA Attribute",
					fmt.Sprintf("Missing PVWA attribute: %s", name))
			}
		}
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
	identityAuthAPI := cybrapi.NewIdentityAuthAPI(fmt.Sprintf(cloudAuthURL, t))

	token, err := identityAuthAPI.GetToken(ctx, cid, []byte(data.ClientSecret.ValueString()))

	if err != nil {
		resp.Diagnostics.AddError("Failed to get authentication token",
			fmt.Sprintf("Failed to get authentication token from Cyberark ISPSS service: %+v", err))
		return
	}

	// Create a client for Cyberark PAM
	pamAPI := cybrapi.NewPAMAPI(fmt.Sprintf(cloudPamURL, d), token, true)

	// Create a client for Cyberark SecretsHub
	secretsHubAPI := cybrapi.NewSecretsHubAPI(fmt.Sprintf(cloudSecretsHubURL, d), token)

	var pvwaAPI cybrapi.PAMAPI = nil
	if data.PVWAURL.ValueString() != "" {
		// Default to "cyberark" login method if not set
		loginMethod := "cyberark"
		if data.PVWALoginMethod.ValueString() != "" {
			loginMethod = data.PVWALoginMethod.ValueString()
		}

		pvwaAuthAPI := cybrapi.NewPVWAAuthAPI(data.PVWAURL.ValueString(), loginMethod)
		pvwaToken, err := pvwaAuthAPI.GetToken(ctx, data.PVWAUsername.ValueString(), []byte(data.PVWAPassword.ValueString()))

		if err != nil {
			resp.Diagnostics.AddError("Failed to get PVWA authentication token",
				fmt.Sprintf("Failed to get authentication token from Cyberark PVWA service: %+v", err))
			return
		}

		pvwaAPI = cybrapi.NewPAMAPI(data.PVWAURL.ValueString(), pvwaToken, false)
	}

	resp.DataSourceData = &cybrapi.API{
		PamAPI:        pamAPI,
		SecretsHubAPI: secretsHubAPI,
		PVWAAPI:       pvwaAPI,
	}
	resp.ResourceData = &cybrapi.API{
		PamAPI:        pamAPI,
		SecretsHubAPI: secretsHubAPI,
		PVWAAPI:       pvwaAPI,
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
		NewPVWAAWSAccountResource,
		NewAWSSecretStoreResource,
		NewAzureAccountResource,
		NewPVWAAzureAccountResource,
		NewAzureSecretStoreResource,
		NewDBAccountResource,
		NewPVWADBAccountResource,
		NewSafeResource,
		NewPVWASafeResource,
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
