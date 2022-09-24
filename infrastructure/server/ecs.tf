# TODO: rclone config
# TODO: check teardown is run - it's not on stop
# TODO: status stuck at pending, because of teardown?
# TODO: port + public IP
# TODO: health?
# TODO: rcon doesn't connect

resource "aws_ecs_task_definition" "task" {
  execution_role_arn       = var.execution_role_arn
  task_role_arn            = var.sidecars_role_arn
  family                   = var.name
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  volume {
    name = "data"
  }
  cpu    = "1024"
  memory = "2048"
  container_definitions = templatefile(
    "${path.module}/task-template.json",
    {
      name              = var.name
      setup_image       = var.setup_image
      backup_image      = var.backup_image
      main_image        = var.main_image
      teardown_image    = var.teardown_image
      bucket_name       = var.bucket_name
      sidecars_role_arn = var.sidecars_role_arn
      region            = var.region
      minecraft_type = var.minecraft_type
      minecraft_version = var.minecraft_version
      forge_version = var.forge_version
    }
  )
}