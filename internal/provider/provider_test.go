package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cyberark/terraform-provider-cyberark/internal/provider"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	providerConfig = testProviderConfigData()
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"cybr-sh": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

func TestProviderResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwprovider.SchemaRequest{}
	schemaResponse := &fwprovider.SchemaResponse{}

	// Instantiate the provider.Provider and call its Schema method
	provider.New("test")().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func testProviderConfigData() string {
	return fmt.Sprintf(`
        provider "cybr-sh" {
            tenant        = %[1]q
            domain        = %[2]q
            client_id     = %[3]q
            client_secret = %[4]q
        }`, os.Getenv("TF_TENANT_NAME"), os.Getenv("TF_DOMAIN_NAME"), os.Getenv("TF_CLIENT_ID"), os.Getenv("TF_CLIENT_SECRET"))
}
