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

variable "safename" {}

terraform {
  required_providers {
    cybr-sh = {
      source = "example/cyberark/cybr-sh"
      version = "~> 0"
    }
  }
}


provider "cybr-sh" {
  tenant       = var.tenant_name
  domain       = var.domain
  client_id     = var.client_id
  client_secret = var.client_secret
}

resource "cybr-sh_safe" "safetesting" {
  safe_name          = var.safename
  safe_desc          = "This is for safe testing"
  member             = "secretshub"
  member_type        = "user"
  permission_level   = "full"
  retention          = 7
  purge              = false
}

output "status" {
  value = (
    cybr-sh_safe.safetesting.id != "" ? "success" : "fail"
  )
}