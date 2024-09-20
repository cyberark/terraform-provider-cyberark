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

variable "aws_store_name" {}
variable "aws_alias" {}
variable "aws_region" {}
variable "aws_account_id" {}
variable "aws_iam_role" {}


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

resource "cyberark_aws_secret_store" "awssecretstorecreation" {
  name              = var.aws_store_name
  description       = "This aws store for created for testing purpose"
  aws_account_alias  = var.aws_alias
  aws_account_id     = var.aws_account_id
  aws_account_region = var.aws_region
  aws_iam_role       = var.aws_iam_role
}

output "status" {
  value = (
    cyberark_aws_secret_store.awssecretstorecreation.id != "" ? "success" : "fail"
  )
}
