variable "tenant_name" {
  description = "CyberArk Shared Services Tenant"
  type        = string
}
variable "client_id" {
  description = "The username for secretshub service account"
  type        = string
}
variable "client_secret" {
  description = "The password for secretshub service account"
  type        = string
  sensitive   = true
}
variable "domain" {
  description = "CyberArk Privilege Cloud and Secrets Hub Domain"
  type        = string
}

variable "secret_key" {
  description = "Secret Key of the credential object"
  type      = string
  sensitive = true
}

variable "safename" {}
variable "aws_username" {}
variable "policy_name" {}
variable "source_p_cloud_id" {}
variable "target_secretstore_id" {}
variable "aws_alias" {}
variable "aws_region" {}
variable "aws_account_id" {}
variable "aws_key_id" {}

terraform {
  required_providers {
    cyberark = {
      source = "example/cyberark/cyberark"
      version = "~> 0"
    }
    time = {
      source = "hashicorp/time"
      version = "0.11.1"
    }
  }
}

provider "cyberark" {
  tenant       = var.tenant_name
  domain       = var.domain
  client_id    = var.client_id
  client_secret = var.client_secret
}

resource "cyberark_safe" "safetesting" {
  safe_name          = var.safename
  safe_desc          = "This is for safe testing"
  member             = "secretshub"
  member_type        = "user"
  permission_level   = "full"
  retention          = 7
  purge              = false
}

resource "time_sleep" "wait_5_seconds" {
  depends_on = [cyberark_safe.safetesting]
  create_duration = "5s"
}

resource "cyberark_aws_account" "awsaccountcreation" {
  name              = var.aws_username
  username          = var.aws_username
  platform          = "AWSAccessKeys"
  safe              = var.safename
  secret            = var.secret_key
  sm_manage         = false
  sm_manage_reason  = "No CPM Associated with Safe."
  aws_kid           = var.aws_key_id
  aws_account_id     = var.aws_account_id
  aws_alias         = var.aws_alias
  secret_name_in_secret_store = "aws_testing"
  depends_on = [time_sleep.wait_5_seconds]
}


resource "time_sleep" "wait_few_seconds" {
  depends_on = [cyberark_aws_account.awsaccountcreation]
  create_duration = "5s"
}

resource "cyberark_sync_policy" "syncpolicycreate" {
  name              = var.policy_name
  description       = "Policy description"
  source_id         = var.source_p_cloud_id
  target_id         = var.target_secretstore_id
  safe_name         = var.safename
  depends_on = [time_sleep.wait_few_seconds]
}

output "overall_status" {
  value = (
    (
      cyberark_safe.safetesting.id != "" &&
      cyberark_aws_account.awsaccountcreation.id != "" &&
      cyberark_sync_policy.syncpolicycreate.id != ""
    ) ? "success" : "fail"
  )
}