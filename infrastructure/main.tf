data "aws_s3_bucket" "minecraft" {
  bucket = var.bucket
}

resource "aws_kms_key" "minecraft" {
  description             = "minecraft"
  deletion_window_in_days = 7
}

resource "aws_cloudwatch_log_group" "ecs-cluster" {
  name = "minecraft"
}

resource "aws_ecs_cluster" "minecraft" {
  name = "minecraft"

  configuration {
    execute_command_configuration {
      kms_key_id = aws_kms_key.minecraft.arn
      logging    = "OVERRIDE"

      log_configuration {
        cloud_watch_encryption_enabled = true
        cloud_watch_log_group_name     = aws_cloudwatch_log_group.ecs-cluster.name
      }
    }
  }
}

resource "aws_ecs_cluster_capacity_providers" "minecraft" {
  cluster_name = aws_ecs_cluster.minecraft.name

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}

resource "aws_api_gateway_rest_api" "minecraft" {
  name = "minecraft"
}

data "aws_iam_policy_document" "allow-from-ip" {
  statement {
    effect = "Allow"
    actions = ["execute-api:Invoke"]
    principals {
      identifiers = ["*"]
      type        = "*"
    }
    resources = ["${aws_api_gateway_rest_api.minecraft.execution_arn}/*/*/*"]
    condition {
      test     = "IpAddress"
      values   = [var.home_ip]
      variable = "aws:SourceIp"
    }
  }
}

resource "aws_api_gateway_rest_api_policy" "storage_list" {
  policy      = data.aws_iam_policy_document.allow-from-ip.json
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
}

data "aws_region" "default" {}
data "aws_caller_identity" "current" {}