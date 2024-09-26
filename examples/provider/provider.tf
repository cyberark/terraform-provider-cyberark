variable "secret_key" {
  type      = string
  sensitive = true
}

provider "cyberark" {
  tenant        = "aarp0000"
  domain        = "example-domain"
  client_id     = "automation@cyberark.cloud.aarp0000"
  client_secret = var.secret_key
}
