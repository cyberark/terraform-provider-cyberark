resource "secretshub_sync_policy" "syncpolicy" {
  name           = "azure_policy"
  description    = "Policy description"
  source_id      = "Source ID"
  target_id      = "Target ID"
  safe_name      = "TF_TEST_SAFE"
  transformation = "password_only_plain_text"
}