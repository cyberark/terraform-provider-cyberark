package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cyberark/terraform-provider-cyberark/internal/provider"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAzureAccountResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	// Instantiate the resource.Resource and call its Schema method
	provider.NewAzureAccountResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestAccAzureAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccAzureAccountCreateData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "name", os.Getenv("TF_AZURE_NAME")),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "address", "1.2.3.4"),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "username", os.Getenv("TF_AZURE_USERNAME")),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "platform", "MS_Azure"),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "safe", "Testsafe"),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "secret", os.Getenv("TF_AZURE_SECRET")),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "sm_manage", "false"),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "sm_manage_reason", "No CPM Associated with Safe"),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "ms_app_id", os.Getenv("TF_AZURE_APP_ID")),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "ms_app_obj_id", os.Getenv("TF_AZURE_OBJ_ID")),
					resource.TestCheckResourceAttr("cyberark_azure_account.test", "ms_key_id", os.Getenv("TF_AZURE_KEY_ID")),
					resource.TestCheckResourceAttrSet("cyberark_azure_account.test", "id"),
					resource.TestCheckResourceAttrSet("cyberark_azure_account.test", "last_updated"),
				),
			},
			{
				Config: providerConfig + `
				  removed {
					from = cyberark_azure_account.test
					lifecycle {
						destroy = false
					}
				
				}`,
			},
		},
	})
}

func testAccAzureAccountCreateData() string {
	return fmt.Sprintf(`
	resource "cyberark_azure_account" "test" {
		name             = %[1]q
		address          = "1.2.3.4"
		username         = %[2]q
		platform         = "MS_Azure"
		safe             = "Testsafe"
		secret           = %[3]q
		sm_manage        = false
		sm_manage_reason = "No CPM Associated with Safe"
		ms_app_id        = %[4]q
		ms_app_obj_id    = %[5]q
		ms_key_id        = %[6]q
		    
}
	`, os.Getenv("TF_AZURE_NAME"), os.Getenv("TF_AZURE_USERNAME"), os.Getenv("TF_AZURE_SECRET"),
		os.Getenv("TF_AZURE_APP_ID"), os.Getenv("TF_AZURE_OBJ_ID"), os.Getenv("TF_AZURE_KEY_ID"))
}
