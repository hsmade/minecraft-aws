resource "aws_efs_file_system" "data" {
  creation_token = var.name
  encrypted = true
  lifecycle_policy {
    transition_to_ia = "AFTER_7_DAYS"
    #    transition_to_primary_storage_class = "NONE"
  }
}
