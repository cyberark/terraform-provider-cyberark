# Import the existing account ID from create stage
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

variable "secret_key" {
  description = "Secret Key of the credential object"
  type        = string
  sensitive   = true
}

variable "safe_name" {
  description = "Name of the safe to update"
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

# Update the safe with new values
resource "cyberark_safe" "imported" {
  safe_name          = var.safe_name
  safe_desc          = "Updated for Safe CRUD testing"  # Changed value
  member             = "secretshub"
  member_type        = "user"
  permission_level   = "full"
  retention_versions = 10
  purge              = false
}

output "update_status" {
  value = (
    cyberark_safe.imported.safe_desc == "Updated for Safe CRUD testing" &&
    cyberark_safe.imported.retention_versions == 10
  ) ? "success" : "fail"
}
