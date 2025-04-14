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

variable "tenant_name" {
  description = "CyberArk Shared Services Tenant"
  type        = string
}

variable "domain" {
  description = "CyberArk Privilege Cloud and Secrets Hub Domain"
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

provider "cyberark" {
  tenant        = var.tenant_name
  domain        = var.domain
  client_id     = var.client_id
  client_secret = var.client_secret
}

variable "secret_key" {
  description = "Secret Key of the credential object"
  type        = string
  sensitive   = true
}

variable "safe_name" {
  description = "Name of the safe to create"
  type        = string
}

variable "aws_username" {
  description = "AWS account username"
  type        = string
}

variable "aws_key_id" {
  description = "AWS Key ID"
  type        = string
}

variable "aws_account_id" {
  description = "AWS Account ID"
  type        = string
}

variable "aws_alias" {
  description = "AWS account alias"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

# First create a safe for the account
resource "cyberark_safe" "test_safe" {
  safe_name        = var.safe_name
  safe_desc        = "Created for AWS account CRUD testing"
  member           = "secretshub"
  member_type      = "user"
  permission_level = "full"
  retention        = 0
  purge            = false
}

resource "time_sleep" "wait_after_safe" {
  depends_on = [cyberark_safe.test_safe]
  create_duration = "5s"
}

# Create the AWS account
resource "cyberark_aws_account" "test_account" {
  name                        = var.aws_username
  username                    = var.aws_username
  platform                    = "AWSAccessKeys"
  safe                        = cyberark_safe.test_safe.safe_name
  secret                      = var.secret_key
  secret_name_in_secret_store = "aws_testing"
  sm_manage                   = false
  sm_manage_reason            = "No CPM Associated with Safe."
  aws_kid                     = var.aws_key_id
  aws_account_id              = var.aws_account_id
  aws_alias                   = var.aws_alias
  aws_account_region          = var.aws_region

  depends_on = [time_sleep.wait_after_safe]
}

# Save the ID for later use in other stages
output "account_id" {
  value = cyberark_aws_account.test_account.id
}

output "safe_name" {
  value = cyberark_safe.test_safe.safe_name
}

output "create_status" {
  value = cyberark_aws_account.test_account.id != "" ? "success" : "fail"
}
