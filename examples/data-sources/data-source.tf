data "secretshub_auth_token" "token" {}

output "ispss_tk" {
  value     = data.secretshub_auth_token.token
  sensitive = true
}
