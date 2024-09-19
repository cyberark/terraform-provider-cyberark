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

func TestSafeResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	// Instantiate the resource.Resource and call its Schema method
	provider.NewSafeResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestSafeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testSafeCreateData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cybr-sh_safe.test", "safe_name", os.Getenv("TF_SAFE_NAME")),
					resource.TestCheckResourceAttr("cybr-sh_safe.test", "safe_desc", "This is for safe testing"),
					resource.TestCheckResourceAttr("cybr-sh_safe.test", "member", "secretshub"),
					resource.TestCheckResourceAttr("cybr-sh_safe.test", "member_type", "user"),
					resource.TestCheckResourceAttr("cybr-sh_safe.test", "permission_level", "full"),
					resource.TestCheckResourceAttr("cybr-sh_safe.test", "retention", "7"),
					resource.TestCheckResourceAttr("cybr-sh_safe.test", "purge", "false"),
					resource.TestCheckResourceAttrSet("cybr-sh_safe.test", "id"),
					resource.TestCheckResourceAttrSet("cybr-sh_safe.test", "last_updated"),
				),
			},
			{
				Config: providerConfig + `
				  removed {
					from = cybr-sh_safe.test
					lifecycle {
						destroy = false
					}
				
				}`,
			},
		},
	})
}

func testSafeCreateData() string {
	return fmt.Sprintf(`
	resource "cybr-sh_safe" "test" {
		safe_name          = %[1]q
		safe_desc          = "This is for safe testing"
		member             = "secretshub"
		member_type        = "user"
		permission_level   = "full"
		retention          = 7
		purge              = false
}
	`, os.Getenv("TF_SAFE_NAME"))
}
