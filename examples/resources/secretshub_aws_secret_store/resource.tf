resource "secretshub_aws_secret_store" "awstest" {
  name               = "aws_store"
  description        = "AWS store for testing purpose"
  aws_account_alias  = "conjurdev"
  aws_account_id     = "AWS Account ID"
  aws_account_region = "us-east-1"
  aws_iam_role       = "AWS IAM Role"
}