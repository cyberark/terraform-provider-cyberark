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

func TestAWSSecretStoreResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := fwresource.SchemaRequest{}
	schemaResponse := &fwresource.SchemaResponse{}

	// Instantiate the resource.Resource and call its Schema method
	provider.NewAWSSecretStoreResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Validate the schema
	diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)

	if diagnostics.HasError() {
		t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
	}
}

func TestAwsSecretStoreResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAWSSecretSyncPolicyCreateData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cybr-sh_aws_secret_store.test", "name", "aws_store"),
					resource.TestCheckResourceAttr("cybr-sh_aws_secret_store.test", "description", "AWS store for testing purpose"),
					resource.TestCheckResourceAttr("cybr-sh_aws_secret_store.test", "aws_account_alias", os.Getenv("TF_AWS_ALIAS")),
					resource.TestCheckResourceAttr("cybr-sh_aws_secret_store.test", "aws_account_id", os.Getenv("TF_AWS_ACCOUNT_ID")),
					resource.TestCheckResourceAttr("cybr-sh_aws_secret_store.test", "aws_account_region", os.Getenv("TF_AWS_ACCOUNT_REGION")),
					resource.TestCheckResourceAttr("cybr-sh_aws_secret_store.test", "aws_iam_role", os.Getenv("TF_AWS_IAM_ROLE")),
					resource.TestCheckResourceAttrSet("cybr-sh_aws_secret_store.test", "id"),
					resource.TestCheckResourceAttrSet("cybr-sh_aws_secret_store.test", "last_updated"),
					resource.TestCheckResourceAttr("cybr-sh_sync_policy.test", "name", "aws_policy"),
					resource.TestCheckResourceAttr("cybr-sh_sync_policy.test", "description", "Policy description"),
					resource.TestCheckResourceAttr("cybr-sh_sync_policy.test", "source_id", os.Getenv("TF_SOURCE_ID")),
					resource.TestCheckResourceAttr("cybr-sh_sync_policy.test", "safe_name", "Testsafe"),
				),
			},
			{
				Config: providerConfig + `
				  removed {
					from = cybr-sh_aws_secret_store.test
					lifecycle {
						destroy = false
					}
				}
				  removed {
					from = cybr-sh_sync_policy.test
					lifecycle {
						destroy = false
				    }
				}`,
			},
		},
	})
}

func testAWSSecretSyncPolicyCreateData() string {
	return fmt.Sprintf(`
    resource "cybr-sh_aws_secret_store" "test" {
                    name               = "aws_store"
                    description        = "AWS store for testing purpose"
                    aws_account_alias  = %[1]q
                    aws_account_id     = %[2]q
                    aws_account_region = %[3]q
                    aws_iam_role       = %[4]q
                }

                resource "cybr-sh_sync_policy" "test" {
                    name           = "aws_policy"
                    description    = "Policy description"
                    source_id      = %[5]q
                    target_id      = cybr-sh_aws_secret_store.test.id
                    safe_name      = "Testsafe"
                    depends_on     = [cybr-sh_aws_secret_store.test]
                }
    `, os.Getenv("TF_AWS_ALIAS"), os.Getenv("TF_AWS_ACCOUNT_ID"), os.Getenv("TF_AWS_ACCOUNT_REGION"), os.Getenv("TF_AWS_IAM_ROLE"),
		os.Getenv("TF_SOURCE_ID"))
}
