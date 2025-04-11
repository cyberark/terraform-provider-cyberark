terraform {
  required_providers {
    cyberark = {
      source  = "example/cyberark/cyberark"
      version = "~> 0"
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

# Update the AWS secret store with new values
resource "cyberark_aws_secret_store" "imported" {
  name               = var.aws_store_name
  description        = "Updated AWS store for CRUD testing" # Changed value
  aws_account_alias  = var.aws_alias
  aws_account_id     = var.aws_account_id
  aws_account_region = var.aws_region
  aws_iam_role       = var.aws_iam_role
}

output "update_status" {
  value = (
    cyberark_aws_secret_store.imported.description == "Updated AWS store for CRUD testing"
  ) ? "success" : "fail"
}
