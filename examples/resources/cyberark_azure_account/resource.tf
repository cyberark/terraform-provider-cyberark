variable "secret_key" {
  type      = string
  sensitive = true
}

resource "cyberark_azure_account" "mskey" {
  name             = "user-ms"
  address          = "1.2.3.4"
  username         = "user-ms"
  platform         = "MS_TF"
  safe             = "TF_TEST_SAFE"
  secret           = var.secret_key
  sm_manage        = false
  sm_manage_reason = "No CPM Associated with Safe."
  ms_app_id        = "Application ID"
  ms_app_obj_id    = "Application Object ID"
  ms_key_id        = "Key ID"
  ms_ad_id         = "AD Key ID"
  ms_duration      = "300"
  ms_pop           = "yes"
  ms_key_desc      = "key descriptiong with spaces"
}
