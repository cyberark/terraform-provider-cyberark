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

func TestDBAccountResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	// Instantiate the resource.Resource and call its Schema method
	provider.NewDBAccountResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestAccDBAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccDBAccountCreateData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "name", os.Getenv("TF_DB_NAME")),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "address", "1.2.3.4"),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "username", os.Getenv("TF_DB_USERNAME")),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "platform", "MySQL"),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "safe", "Testsafe"),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "secret", os.Getenv("TF_DB_SECRET")),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "secret_name_in_secret_store", "user"),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "sm_manage", "false"),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "sm_manage_reason", "No CPM Associated with Safe"),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "db_port", "8432"),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "db_dsn", "dsn"),
					resource.TestCheckResourceAttr("cybr-sh_db_account.test", "dbname", "dbo.services"),
					resource.TestCheckResourceAttrSet("cybr-sh_db_account.test", "id"),
					resource.TestCheckResourceAttrSet("cybr-sh_db_account.test", "last_updated"),
				),
			},
			{
				Config: providerConfig + `
				  removed {
					from = cybr-sh_db_account.test
					lifecycle {
						destroy = false
					}
				
				}`,
			},
		},
	})
}

func testAccDBAccountCreateData() string {
	return fmt.Sprintf(`
	resource "cybr-sh_db_account" "test" {
		name                        = %[1]q
		address                     = "1.2.3.4"
		username                    = %[2]q
		platform                    = "MySQL"
		safe                        = "Testsafe"
		secret                      = %[3]q
		secret_name_in_secret_store = "user"
		sm_manage                   = false
		sm_manage_reason            = "No CPM Associated with Safe"
		db_port                     = "8432"
		db_dsn                      = "dsn"
		dbname                      = "dbo.services"
}
	`, os.Getenv("TF_DB_NAME"), os.Getenv("TF_DB_USERNAME"), os.Getenv("TF_DB_SECRET"))
}
