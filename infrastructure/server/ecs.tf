# task family
# task definition
# FIXME: rclone config
# FIXME: IAM for S3

#resource "aws_ecs_task_definition" "task" {
#  family = var.name
#  volume {
#    name = "data"
#  }
#  cpu = "1024"
#  memory = "2048"
#  container_definitions = jsonencode([
#    {
#      # copies the S3 to the volume
#      name = "setup"
#      image = var.setup_image
#      command = "/scripts/setup.sh"
#      essential = false
#      mountPoints = [
#        {
#          containerPath = "/data"
#          sourceVolume = "data"
#        }
#      ]
#      environment = [
#        {
#          name = "url"
#          value = var.bucket_url
#        }
#      ]
#    },
#    {
#      # backup
#      name = "backup"
#      image = var.backup_image
#      essential = false
#      mountPoints = [
#        {
#          containerPath = "/data"
#          sourceVolume = "data"
#        }
#      ]
#      environment = [
#        {
#          name = "RCON_HOST"
#          value = "main"
#        },
#        {
#          name = "BACKUP_METHOD"
#          value = "rclone"
#        },
#        {
#          name = "BACKUP_NAME"
#          value = var.name
#        },
#        {
#          name = "INITIAL_DELAY"
#          value = "5m"
#        },
#        {
#          name = "BACKUP_INTERVAL"
#          value = "5m"
#        },
#        {
#          name = "PAUSE_IF_NO_PLAYERS"
#          value = "true"
#        },
#        {
#          name = "PLAYERS_ONLINE_CHECK_INTERVAL"
#          value = "5m"
#        },
#        {
#          name = "PRUNE_BACKUPS_DAYS"
#          value = "3"
#        },
#        {
#          name = "EXCLUDES"
#          value = "cache,logs"
#        },
#        {
#          name = "RCLONE_REMOTE"
#          value = "s3"
#        }
#      ]
#      # FIXME: /config/rclone/rclone.conf
#    },
#    {
#      name = "main"
#      image = var.main_image
#      essential = true
#      mountPoints = [
#        {
#          containerPath = "/data"
#          sourceVolume = "data"
#        }
#      ]
#      environment = [
#        {
#          name = "EULA"
#          value = "TRUE"
#        }
#      ]
#      dependsOn = [
#        {
#          containerName = "setup"
#          condition = "COMPLETE"
#        }
#      ]
#    },
#    {
#      name = "teardown"
#      image = var.teardown_image
#      command = "/scripts/teardown.sh"
#      essential = false
#      mountPoints = [
#        {
#          containerPath = "/data"
#          sourceVolume = "data"
#        }
#      ]
#      dependsOn = [
#        {
#          containerName = "main"
#          condition = "COMPLETE"
#        }
#      ]
#    }
#  ])
#}