package provider_test

import (
	"context"
	"github.com/cyberark/terraform-provider-secretshub/internal/provider"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
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
