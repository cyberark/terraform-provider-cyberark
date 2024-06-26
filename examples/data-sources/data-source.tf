data "cybr-sh_auth_token" "token" {}

output "ispss_tk" {
  value     = data.cybr-sh_auth_token.token
  sensitive = true
}
