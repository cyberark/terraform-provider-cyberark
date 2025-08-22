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

variable "safe_name" {
  description = "Name of the safe to create"
  type        = string
}

# Create the safe
resource "cyberark_safe" "test_safe" {
  safe_name        = var.safe_name
  safe_desc        = "Created for Safe CRUD testing"
  member           = "secretshub"
  member_type      = "user"
  permission_level = "full"
  # Uses the default retention policy (RetentionDays = 7)
}

# Save the ID for later use in other stages
output "safe_name" {
  value = cyberark_safe.test_safe.safe_name
}

output "create_status" {
  value = cyberark_safe.test_safe.safe_name != "" ? "success" : "fail"
}
