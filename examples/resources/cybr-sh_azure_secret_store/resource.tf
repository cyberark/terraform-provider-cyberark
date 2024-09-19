variable "azure_app_secret" {
  type      = string
  sensitive = true
}

resource "cybr-sh_azure_secret_store" "storecreation" {
  name                          = "azure_secret_store"
  description                   = "AKV Secrets Manager for dev-team"
  azure_app_client_directory_id = "Azure App Client Directory ID"
  azure_vault_url               = "Azure Vault URL"
  azure_app_client_id           = "Azure App Client ID"
  azure_app_client_secret       = var.azure_app_secret
  connection_type               = "CONNECTOR"
  connector_id                  = "Connector ID"
  subscription_id               = "Subscription ID"
  subscription_name             = "Subscription Name"
  resource_group_name           = "test_group"
}
