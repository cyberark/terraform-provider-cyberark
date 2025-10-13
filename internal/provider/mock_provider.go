package provider

import (
	"context"
	cybrapi "github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Returns a ProtoV6 provider factory for mock provider, injecting the given API.
func mockProtoV6ProviderFactories(api *cybrapi.API) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"cyberark": providerserver.NewProtocol6WithError(
			mockProvider{
				realProvider: New("test")(),
				api:          api,
			}),
	}
}

type mockProvider struct {
	realProvider fwprovider.Provider
	api          *cybrapi.API
}

func (m mockProvider) Metadata(ctx context.Context, request fwprovider.MetadataRequest, response *fwprovider.MetadataResponse) {
	m.realProvider.Metadata(ctx, request, response)
}

func (m mockProvider) Schema(context.Context, fwprovider.SchemaRequest, *fwprovider.SchemaResponse) {
}

func (m mockProvider) Configure(context.Context, fwprovider.ConfigureRequest, *fwprovider.ConfigureResponse) {
}

func (m mockProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}

func (m mockProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newMockPVWASafeObjectResourceFunc(m.api),
	}
}

func newMockPVWASafeObjectResourceFunc(api *cybrapi.API) func() resource.Resource {
	return func() resource.Resource {
		return NewPVWASafeResourceWithAPI(api)
	}
}
