# Import the existing sync policy from create stage
# This would use terraform import in the test script

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

# Create a different AWS secret store resource
resource "cyberark_aws_secret_store" "another_test_store" {
  name               = var.aws_store_name
  description        = "A different store created for Sync Policy CRUD testing"
  aws_account_alias  = var.aws_alias
  aws_account_id     = var.aws_account_id
  aws_account_region = var.aws_region
  aws_iam_role       = var.aws_iam_role
}

resource "time_sleep" "wait_after_secret_store" {
  depends_on = [cyberark_aws_secret_store.another_test_store]
  create_duration = "5s"
}

# Update the Sync Policy with new values
resource "cyberark_sync_policy" "imported" {
  name        = var.policy_name
  description = "Updated Sync Policy for testing"
  safe_name   = var.safe_name
  source_id   = var.source_p_cloud_id
  target_id   = cyberark_aws_secret_store.another_test_store.id

  depends_on = [time_sleep.wait_after_secret_store]
}

resource "time_sleep" "wait_after_sync_policy" {
  depends_on = [cyberark_sync_policy.imported]
  create_duration = "20s"
}

output "update_status" {
  value = (
    cyberark_sync_policy.imported.description == "Updated Sync Policy for testing" &&
    cyberark_sync_policy.imported.target_id == cyberark_aws_secret_store.another_test_store.id
  ) ? "success" : "fail"
}