data "cyberark_auth_token" "token" {}

output "ispss_tk" {
  value     = data.cyberark_auth_token.token
  sensitive = true
}
