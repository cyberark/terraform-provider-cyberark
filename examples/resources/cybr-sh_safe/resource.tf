resource "cybr-sh_safe" "AAM_Test_Safe" {
  safe_name          = "GEN_BY_TF_abc"
  safe_desc          = "Description for GEN_BY_TF_abc"
  member             = "demo@cyberark.cloud.aarp0000"
  member_type        = "user"
  permission_level   = "full" # full, read, approver, manager
  retention          = 7
  retention_versions = 7
  purge              = false
  cpm_name           = "PasswordManager"
  safe_loc           = ""
}
