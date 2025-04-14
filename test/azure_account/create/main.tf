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

# First create a safe for the account
resource "cyberark_safe" "test_safe" {
  safe_name        = var.safe_name
  safe_desc        = "Created for Azure account CRUD testing"
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

# Create the Azure account
resource "cyberark_azure_account" "test_account" {
  name                        = var.azure_username
  username                    = var.azure_username
  platform                    = "MS_Azure"
  safe                        = cyberark_safe.test_safe.safe_name
  secret                      = var.secret_key
  secret_name_in_secret_store = "azure_testing"
  sm_manage                   = false
  sm_manage_reason            = "No CPM Associated with Safe."
  ms_app_id                   = var.ms_app_id
  ms_app_obj_id               = var.ms_app_obj_id
  ms_key_id                   = var.ms_key_id

  depends_on = [time_sleep.wait_after_safe]
}

# Save the ID for later use in other stages
output "account_id" {
  value = cyberark_azure_account.test_account.id
}

output "safe_name" {
  value = cyberark_safe.test_safe.safe_name
}

output "create_status" {
  value = cyberark_azure_account.test_account.id != "" ? "success" : "fail"
}
