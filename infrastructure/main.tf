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

data "aws_vpc" "vpc" {}

resource "aws_security_group" "minecraft" {
  name = "minecraft"

  ingress {
    description = "minecraft"
    from_port = 25565
    to_port   = 25565
    protocol  = "TCP"
    cidr_blocks = [
      "${var.home_ip}/32"  # FIXME: how to open to others
    ]
  }

  ingress {
    description = "metrics"
    from_port = 8080
    to_port   = 8080
    protocol  = "TCP"
    cidr_blocks = [
      "${var.home_ip}/32"
    ]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "ALL"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "efs" {
  name = "minecraft-efs"
}

# allow ecs to connect to efs
resource "aws_security_group_rule" "nfs" {
  from_port         = 2049
  protocol          = "TCP"
  security_group_id = aws_security_group.efs.id
  source_security_group_id = aws_security_group.minecraft.id
  to_port           = 2049
  type              = "ingress"
}

#resource "aws_iam_role" "ssm" {
#  assume_role_policy = ""
#}
#
#resource "aws_backup_plan" "efs" {
#  name = "minecraft EFS daily backup"
#  rule {
#    rule_name         = "minecraft_efs_daily_backup"
#    target_vault_name = "minecraft"
#    schedule = "cront(0 0 * * ? *)"
#
#    lifecycle {
#      delete_after = 14
#    }
#  }
#
#  advanced_backup_setting {
#    resource_type = "EFS"
#    backup_options = {}
#  }
#}
# FIXME: missing
# iam role for ssm + efs
# backup policy for EFS