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

variable "safe_name" {
  description = "Name of the safe to create"
  type        = string
}

variable "aws_store_name" {
  description = "AWS secret store name"
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

variable "aws_iam_role" {
  description = "AWS IAM role"
  type        = string
}

variable "policy_name" {
  description = "Name of the sync policy to create"
  type = string
}

variable "source_p_cloud_id" {
  description = "Source Privilege Cloud ID"
  type = string
}

# First create a safe for the sync policy
resource "cyberark_safe" "test_safe" {
  safe_name        = var.safe_name
  safe_desc        = "Created for Sync Policy CRUD testing"
  member           = "secretshub"
  member_type      = "user"
  permission_level = "full"
  retention        = 0
  purge            = false
}

# Create AWS secret store resource
resource "cyberark_aws_secret_store" "test_store" {
  name               = var.aws_store_name
  description        = "Created for Sync Policy CRUD testing"
  aws_account_alias  = var.aws_alias
  aws_account_id     = var.aws_account_id
  aws_account_region = var.aws_region
  aws_iam_role       = var.aws_iam_role
}

resource "time_sleep" "wait_after_safe" {
  depends_on = [cyberark_safe.test_safe]
  create_duration = "5s"
}

resource "time_sleep" "wait_after_secret_store" {
  depends_on = [cyberark_aws_secret_store.test_store]
  create_duration = "5s"
}

# Create the Sync Policy
resource "cyberark_sync_policy" "sync_policy" {
  name        = var.policy_name
  description = "Created for Sync Policy CRUD testing"
  safe_name   = cyberark_safe.test_safe.safe_name
  source_id   = var.source_p_cloud_id
  target_id   = cyberark_aws_secret_store.test_store.id

  depends_on = [time_sleep.wait_after_safe, time_sleep.wait_after_secret_store]
}

# Save the ID for later use in other stages
output "policy_id" {
  value = cyberark_sync_policy.sync_policy.id
}

output "safe_name" {
  value = cyberark_safe.test_safe.safe_name
}

output "target_id" {
  value = cyberark_aws_secret_store.test_store.id
}

output "create_status" {
  value = cyberark_sync_policy.sync_policy.id != "" ? "success" : "fail"
}
