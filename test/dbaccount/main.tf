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

variable "secret_key" {
  type      = string
  sensitive = true
}
variable "db_username" {}
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

resource "cybr-sh_db_account" "dbcreation" {
  name                        = var.db_username
  address                     = "1.2.3.4"
  username                    = var.db_username
  platform                    = "MySQL"
  safe                        = var.safename
  secret                      = var.secret_key
  secret_name_in_secret_store = "user"
  sm_manage                   = false
  sm_manage_reason            = "No CPM Associated with Safe."
  db_port                     = "8432"
  db_dsn                      = "dsn"
  dbname                      = "dbo.services"
}
 
output "status" {
  value = (
    cybr-sh_db_account.dbcreation.id != "" ? "success" : "fail"
  )
}