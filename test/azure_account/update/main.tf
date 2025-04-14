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

variable "azure_username" {
  description = "Azure account username"
  type        = string
}

variable "ms_app_id" {
  description = "Azure Application ID"
  type        = string
}

variable "ms_app_obj_id" {
  description = "Azure Application Object ID"
  type        = string
}

variable "ms_key_id" {
  description = "Azure Key ID"
  type        = string
}

# Update the Azure account with new values
resource "cyberark_azure_account" "imported" {
  name                        = var.azure_username
  address                     = "1.2.3.4"                # Changed value
  username                    = var.azure_username
  platform                    = "MS_Azure"
  safe                        = var.safe_name
  secret                      = var.secret_key
  secret_name_in_secret_store = "updated_azure_testing"  # Changed value
  sm_manage                   = true                     # Changed value
  sm_manage_reason            = "Updated reason"         # Changed value
  ms_app_id                   = var.ms_app_id
  ms_app_obj_id               = var.ms_app_obj_id
  ms_key_id                   = var.ms_key_id
}

output "update_status" {
  value = (
    cyberark_azure_account.imported.secret_name_in_secret_store == "updated_azure_testing" &&
    cyberark_azure_account.imported.sm_manage == true &&
    cyberark_azure_account.imported.sm_manage_reason == "Updated reason"
  ) ? "success" : "fail"
}
