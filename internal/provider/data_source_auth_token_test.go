package provider_test

import (
	"context"
	"github.com/cyberark/terraform-provider-secretshub/internal/provider"
	"testing"

	fwdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestTokenDataSourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwdatasource.SchemaRequest{}
	schemaResponse := &fwdatasource.SchemaResponse{}

	// Instantiate the datasource.DataSource and call its Schema method
	provider.NewTokenDataSource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}
