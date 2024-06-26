variable "secret_key" {
  type      = string
  sensitive = true
}

resource "cybr-sh_awsaccount" "awskey" {
  name               = "user-aws"
  username           = "user-aws"
  platform           = "AWS_TF"
  safe               = "TF_TEST_SAFE"
  secret             = var.secret_key
  sm_manage          = false
  sm_manage_reason   = "No CPM Associated with Safe."
  aws_kid            = "9876543210"
  aws_account_id     = "0123456789"
  aws_alias          = "aws_alias"
  aws_account_region = "us-east-2"
}
