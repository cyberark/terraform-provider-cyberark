variable "secret_key" {
  type      = string
  sensitive = true
}

resource "cybr-sh_db_account" "pgdb" {
  name                        = "user-db"
  address                     = "1.2.3.4"
  username                    = "user-db"
  platform                    = "MySQL"
  safe                        = "TF_TEST_SAFE"
  secret                      = var.secret_key
  secret_name_in_secret_store = "user"
  sm_manage                   = false
  sm_manage_reason            = "No CPM Associated with Safe."
  db_port                     = "8432"
  db_dsn                      = "dsn"
  dbname                      = "dbo.services"
}
