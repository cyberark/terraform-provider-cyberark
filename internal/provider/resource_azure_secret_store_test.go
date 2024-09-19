package provider_test

import (
	"context"
	"testing"

	"github.com/cyberark/terraform-provider-cyberark/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestAzureSecretStoreResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	// Instantiate the resource.Resource and call its Schema method
	provider.NewAzureSecretStoreResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestAzureSecretStoreResourceUpdate(t *testing.T) {
	t.Parallel()

	// Create a new instance of the azureSecretStoreResource
	r := provider.NewAzureSecretStoreResource()

	// Prepare a context and request/response objects
	ctx := context.Background()
	req := resource.UpdateRequest{}
	var resp resource.UpdateResponse

	// Call the Update method
	r.Update(ctx, req, &resp) // No return value expected

	// Check for errors in the response diagnostics
	if len(resp.Diagnostics) == 0 {
		t.Fatalf("Expected an error diagnostic, but found none")
	}

	// Validate the error message
	expectedError := "Update is not supported through terraform"
	for _, diag := range resp.Diagnostics {
		if diag.Summary() != expectedError {
			t.Fatalf("Expected error message: %s, but got: %s", expectedError, diag.Summary())
		}
	}
}

func TestAzureSecretStoreResourceDelete(t *testing.T) {
	t.Parallel()

	// Create a new instance of the azureSecretStoreResource
	r := provider.NewAzureSecretStoreResource()

	// Prepare a context and request/response objects
	ctx := context.Background()
	req := resource.DeleteRequest{}
	var resp resource.DeleteResponse

	// Call the Delete method
	r.Delete(ctx, req, &resp) // No return value expected

	// Check for errors in the response diagnostics
	if len(resp.Diagnostics) == 0 {
		t.Fatalf("Expected an error diagnostic, but found none")
	}

	// Validate the error message
	expectedError := "Delete is not supported through terraform"
	for _, diag := range resp.Diagnostics {
		if diag.Summary() != expectedError {
			t.Fatalf("Expected error message: %s, but got: %s", expectedError, diag.Summary())
		}
	}
}
