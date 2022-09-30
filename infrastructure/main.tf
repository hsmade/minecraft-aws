resource "aws_cloudwatch_log_group" "ecs-cluster" {
  name = "minecraft"
  retention_in_days = 14
}

resource "aws_ecs_cluster" "minecraft" {
  name = "minecraft"

  configuration {
    execute_command_configuration {
      logging = "OVERRIDE"

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
    base              = 1
    weight            = 100
  }
}

data "aws_iam_policy_document" "ecs_assume_role" {
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

resource "aws_iam_role" "ecs_sidecars" {
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
  name               = "ecs_sidecars"
}

resource "aws_iam_role_policy_attachment" "backup_restore" {
  policy_arn = aws_iam_policy.ecs_backup_restore.arn
  role       = aws_iam_role.ecs_sidecars.name
}

resource "aws_iam_role" "ecs_execution_role" {
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
  name               = "ecs_execution_role"
}

data "aws_iam_policy_document" "ecs_execution_role_rules" {
  statement {
    effect    = "Allow"
    resources = ["*"]
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:CreateLogGroup",
      "kms:GetPublicKey",
      "kms:Decrypt",
      "kms:GenerateDataKey",
      "kms:DescribeKey",
    ]
  }
}

resource "aws_iam_policy" "ecs_execution_role_rules" {
  name   = "ecs_execution_role_rules"
  policy = data.aws_iam_policy_document.ecs_execution_role_rules.json
}

resource "aws_iam_role_policy_attachment" "allow_ecr" {
  policy_arn = aws_iam_policy.ecs_execution_role_rules.arn
  role       = aws_iam_role.ecs_execution_role.name
}

resource "aws_route53_zone" "domain" {
  name = var.domain_name
}

#resource "aws_route53_record" "soa" {
#  zone_id         = aws_route53_zone.domain.id
#  name            = var.domain_name
#  type            = "SOA"
#  ttl             = 30
#  records         = aws_route53_zone.domain.name_servers
#  allow_overwrite = true
#}

data "aws_vpc" "main" {}
data "aws_subnets" "subnets" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.main.id]
  }
}
