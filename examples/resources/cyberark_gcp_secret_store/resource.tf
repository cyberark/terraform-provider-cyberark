resource "cyberark_gcp_secret_store" "gcptest" {
  name                          = "gcp_store"
  description                   = "GCP store for testing purpose"
  gcp_project_name              = "cybr-gcp-project"
  gcp_project_number            = "123456789111213"
  gcp_pool_provider_id          = "secretshub-provider"
  gcp_workload_identity_pool_id = "secretshub-allow-pool"
  service_account_email         = "secretshub-allow-user@cybr-gcp-project-xxxx.iam.gserviceaccount.com"
}