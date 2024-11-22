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
  description = "Secret Key of the credential object"
  type      = string
  sensitive = true
}

variable "aws_username" {}
variable "safename" {}

variable "aws_alias" {}
variable "aws_region" {}
variable "aws_account_id" {}
variable "aws_key_id" {}

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

resource "cyberark_aws_account" "awsaccountcreation" {
  name              = var.aws_username
  username          = var.aws_username
  platform          = "AWSAccessKeys"
  safe              = var.safename
  secret            = var.secret_key
  sm_manage         = false
  sm_manage_reason  = "No CPM Associated with Safe."
  aws_kid           = var.aws_key_id
  aws_account_id     = var.aws_account_id
  aws_alias         = var.aws_alias
  secret_name_in_secret_store = "aws_testing"
}

output "status" {
  value = (
    cyberark_aws_account.awsaccountcreation.id != "" ? "success" : "fail"
  )
}