variable "tenant_name" {}
variable "client_id" {
  description = "The username for secretshub service account"
  type        = string
  sensitive   = true
}
variable "client_secret" {
  description = "The password for secretshub service account"
  type        = string
  sensitive   = true
}
variable "domain" {}
variable "policy_name" {}
variable "source_P_cloud_id" {}
variable "target_secretstore_id" {}
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

resource "cybr-sh_sync_policy" "syncpolicycreate" {
  name              = var.policy_name
  description       = "Policy description"
  source_id         = var.source_P_cloud_id
  target_id         = var.target_secretstore_id
  safe_name         = var.safename
}

output "status" {
  value = (
    cybr-sh_sync_policy.syncpolicycreate.id != "" ? "success" : "fail"
  )
}