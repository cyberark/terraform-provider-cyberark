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

# Create AWS secret store resource
resource "cyberark_aws_secret_store" "test_store" {
  name               = var.aws_store_name
  description        = "AWS store for CRUD testing"
  aws_account_alias  = var.aws_alias
  aws_account_id     = var.aws_account_id
  aws_account_region = var.aws_region
  aws_iam_role       = var.aws_iam_role
}

# Save the ID for later use in other stages
output "store_id" {
  value = cyberark_aws_secret_store.test_store.id
}

output "create_status" {
  value = cyberark_aws_secret_store.test_store.id != "" ? "success" : "fail"
}
