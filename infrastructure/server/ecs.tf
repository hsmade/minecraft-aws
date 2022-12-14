resource "aws_ecs_task_definition" "task" {
  execution_role_arn = var.execution_role_arn
  task_role_arn      = var.sidecars_role_arn
  family             = var.name
  tags = merge({
    "minecraft-version" : var.minecraft_version,
    "forge-version" : var.forge_version,
    "server-type" : var.minecraft_type,
  }, var.tags)
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  volume {
    name = "data"
    efs_volume_configuration {
      file_system_id     = aws_efs_file_system.data.id
      transit_encryption = "ENABLED"
    }
  }
  cpu    = "1024"
  memory = "2048"
  container_definitions = templatefile(
    "${path.module}/task-template.json",
    {
      name              = var.name
      backup_image      = var.backup_image
      main_image        = var.main_image
      rcon_image        = var.rcon_image
      bucket_name       = var.bucket_name
      bucket_region     = var.region
      sidecars_role_arn = var.sidecars_role_arn
      region            = var.region
      minecraft_type    = var.minecraft_type
      minecraft_version = var.minecraft_version
      forge_version     = var.forge_version
      whitelist         = join(",", var.whitelist)
      ops               = join(",", var.ops)
    }
  )
}
