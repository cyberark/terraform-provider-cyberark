// Package provider implements the SecretHub provider for Terraform.
package provider

import (
	"context"
	"fmt"

	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/diag"

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
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "CyberArk Client ID, formatted as username@cyberark.cloud.tenant.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "CyberArk Client ID Password.",
				Optional:    true,
				Sensitive:   true,
			},
			"domain": schema.StringAttribute{
				Description: "CyberArk Privilege Cloud Domain.",
				Optional:    true,
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

// Configure parses the configuration data and initializes the provider.
func (p *secretsHubProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data secretsHubProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the provider configuration data here
	// because in the ValidateConfig method env variables (e.g. client_secret, pvwa_password) are not yet resolved
	resp.Diagnostics.Append(validateProviderConfigModel(data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pamAPI cybrapi.PAMAPI
	var secretsHubAPI cybrapi.SecretsHubAPI

	tenant := data.Tenant.ValueString()
	if tenant != "" {
		clientID := data.ClientID.ValueString()
		domain := data.Domain.ValueString()
		// Create a client for Cyberark ISPSS (Identity Security Platform Shared Services)
		identityAuthAPI := cybrapi.NewIdentityAuthAPI(fmt.Sprintf(cloudAuthURL, tenant))

		token, err := identityAuthAPI.GetToken(ctx, clientID, []byte(data.ClientSecret.ValueString()))

		if err != nil {
			resp.Diagnostics.AddError("Failed to get authentication token",
				fmt.Sprintf("Failed to get authentication token from Cyberark ISPSS service: %+v", err))
			return
		}

		// Create a client for Cyberark PAM
		pamAPI = cybrapi.NewPAMAPI(fmt.Sprintf(cloudPamURL, domain), token, true)

		// Create a client for Cyberark SecretsHub
		secretsHubAPI = cybrapi.NewSecretsHubAPI(fmt.Sprintf(cloudSecretsHubURL, domain), token)
	}

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
		NewSecretStoreStateResource,
		NewGcpSecretStoreResource,
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

func validateProviderConfigModel(data secretsHubProviderModel) (diags diag.Diagnostics) {
	// Validate mandatory pCloud attributes
	if err := validateRequiredAttributes(map[string]string{
		"tenant":        data.Tenant.ValueString(),
		"client_id":     data.ClientID.ValueString(),
		"client_secret": data.ClientSecret.ValueString(),
		"domain":        data.Domain.ValueString(),
	}); err != nil {
		for _, e := range err {
			diags.AddError("Privilege Cloud Vault provider attribute error", e)
		}
	}

	pvwaErrorSummary := "PAM Self-Hosted Vault (PVWA) provider attribute error"
	if err := validatePVWALoginMethod(data.PVWALoginMethod.ValueString()); err != nil {
		diags.AddError(pvwaErrorSummary, err.Error())
	}

	// Validate mandatory PAM self-hosted attributes
	if err := validateRequiredAttributes(map[string]string{
		"pvwa_username": data.PVWAUsername.ValueString(),
		"pvwa_password": data.PVWAPassword.ValueString(),
		"pvwa_url":      data.PVWAURL.ValueString(),
	}); err != nil {
		for _, e := range err {
			diags.AddError(pvwaErrorSummary, e)
		}
	}

	if diags.HasError() {
		return
	}

	if data.Tenant.ValueString() == "" && data.PVWAURL.ValueString() == "" {
		diags.AddError(
			"Missing Configuration",
			"At least one of the following sets of attributes must be provided:\n"+
				"1) tenant, client_id, client_secret, domain\n"+
				"2) pvwa_username, pvwa_password, pvwa_url",
		)
	}
	return
}

func validateRequiredAttributes(attrs map[string]string) []string {
	anySet := false
	for _, v := range attrs {
		if v != "" {
			anySet = true
			break
		}
	}
	var missing []string
	if anySet {
		for name, v := range attrs {
			if v == "" {
				missing = append(missing,
					fmt.Sprintf("Missing required attribute: %s", name),
				)
			}
		}
		return missing
	}
	return nil
}

func validatePVWALoginMethod(method string) error {
	if method == "" {
		//defaults to cyberark
		return nil
	}
	validMethods := map[string]struct{}{
		"cyberark": {},
		"ldap":     {},
		"windows":  {},
		"radius":   {},
	}
	if _, ok := validMethods[method]; !ok {
		return fmt.Errorf("login method (%s) is invalid. Acceptable values: cyberark, ldap, windows, radius", method)
	}
	return nil
}
