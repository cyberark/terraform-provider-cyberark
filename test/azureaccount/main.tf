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
  description = "Password of the credential object"
  type      = string
  sensitive = true
}

variable "azure_username" {}
variable "safename" {}
variable "ms_app_id" {}
variable "ms_app_obj_id" {}
variable "ms_key_id" {}


terraform {
  required_providers {
    cyberark = {
      source = "example/cyberark/cyberark"
      version = "~> 0"
    }
  }
}


provider "cyberark" {
  tenant       = var.tenant_name
  domain       = var.domain
  client_id     = var.client_id
  client_secret = var.client_secret
}

resource "cyberark_azure_account" "msaccountcreation" {
  name             = var.azure_username
  address          = "1.3.3.1"
  username         = var.azure_username
  platform         = "MS_Azure"
  safe             = var.safename
  secret           = var.secret_key
  sm_manage        = false
  sm_manage_reason = "No CPM Associated with Safe."
  ms_app_id         = var.ms_app_id 
  ms_app_obj_id     = var.ms_app_obj_id 
  ms_key_id         = var.ms_key_id
}


output "status" {
  value = (
    cyberark_azure_account.msaccountcreation.id != "" ? "success" : "fail"
  )
}
