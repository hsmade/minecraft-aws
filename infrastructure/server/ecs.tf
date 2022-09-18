# TODO: rclone config
# TODO: test IAM for S3
# TODO: logging

resource "aws_ecs_task_definition" "task" {
  execution_role_arn = var.task_role_arn
  family = var.name
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  volume {
    name = "data"
  }
  cpu    = "1024"
  memory = "2048"
  container_definitions = templatefile(
    "${path.module}//task-template.json",
    {
      name           = var.name
      setup_image    = var.setup_image
      backup_image   = var.backup_image
      main_image     = var.main_image
      teardown_image = var.teardown_image
      bucket_url     = var.bucket_url
      task_role_arn  = var.sidecars_role_arn
    }
  )
}