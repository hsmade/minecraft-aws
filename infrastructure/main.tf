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
    effect  = "Allow"
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

resource "aws_ecs_cluster_capacity_providers" "fargate" {
  cluster_name       = aws_ecs_cluster.minecraft.name
  capacity_providers = ["FARGATE"]
  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    base              = 0
    weight            = 100
  }
}

resource "aws_ecr_repository" "scripts" {
  name = "scripts"
}

data "aws_iam_policy_document" "sidecars" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
    condition {
      test     = "ArnLike"
      values   = ["arn:aws:ecs:${data.aws_region.default.name}:${data.aws_caller_identity.current.account_id}:*"]
      variable = "aws:SourceArn"
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
  }
}

resource "aws_iam_role" "sidecars" {
  assume_role_policy = data.aws_iam_policy_document.sidecars.json
}

data "aws_iam_policy_document" "backup_restore" {
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:GetObjectVersion",
      "s3:DeleteObject",
      "s3:DeleteObjectVersion",
    ]
    resources = ["arn:aws:s3:::${var.bucket}:/*"]
  }
}

resource "aws_iam_policy" "backup_restore" {
  policy = data.aws_iam_policy_document.backup_restore.json
  name   = "backup_restore"
}

resource "aws_iam_role_policy_attachment" "backup_restore" {
  policy_arn = aws_iam_policy.backup_restore.arn
  role       = aws_iam_role.sidecars.name
}

data "aws_iam_policy_document" "task_role" {
  statement {
    effect = "Allow"
    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
    actions = ["sts:AssumeRole"]
  }
}
resource "aws_iam_role" "task_role" {
  assume_role_policy = data.aws_iam_policy_document.task_role.json
}

data "aws_iam_policy_document" "allow_ecr" {
  statement {
    effect = "Allow"
    resources = ["*"]
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
  }
}

# FIXME: is generic
resource "aws_iam_policy" "allow_ecr" {
  name = "allow-ecr"
  policy = data.aws_iam_policy_document.allow_ecr.json
}

resource "aws_iam_role_policy_attachment" "allow_ecr" {
  policy_arn = aws_iam_policy.allow_ecr.arn
  role       = aws_iam_role.task_role.name
}
