resource "aws_efs_file_system" "data" {
  tags = {
    "Name" : var.name
  }
  creation_token = var.name
  encrypted      = true
  lifecycle_policy {
    transition_to_ia = "AFTER_7_DAYS"
    #    transition_to_primary_storage_class = "NONE"
  }

}

resource "aws_efs_mount_target" "data" {
  for_each       = toset(var.subnets)
  file_system_id = aws_efs_file_system.data.id
  subnet_id      = each.value
}
