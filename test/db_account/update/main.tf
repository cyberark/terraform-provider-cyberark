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

variable "db_username" {
  description = "DB account username"
  type        = string
}

# Update the DB account with new values
resource "cyberark_db_account" "imported" {
  name                        = var.db_username
  address                     = "1.2.3.4"
  username                    = var.db_username
  platform                    = "MySQL"
  safe                        = var.safe_name
  secret                      = var.secret_key
  secret_name_in_secret_store = "updated_db_testing"  # Changed value
  sm_manage                   = true                  # Changed value
  sm_manage_reason            = "Updated reason"      # Changed value
  db_port                     = "3306"
  db_dsn                      = "mysql_database"
  dbname                      = "test_db"
}

output "update_status" {
  value = (
    cyberark_db_account.imported.secret_name_in_secret_store == "updated_db_testing" &&
    cyberark_db_account.imported.sm_manage == true &&
    cyberark_db_account.imported.sm_manage_reason == "Updated reason"
  ) ? "success" : "fail"
}
