data "external" "git_checkout" {
  program = ["${path.module}/get_sha.sh"]
}

data "aws_region" "default" {}
data "aws_caller_identity" "current" {}

resource "aws_api_gateway_deployment" "deployment" {
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
  triggers = {
    redeployment = data.external.git_checkout.result.sha
  }
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "stage" {
  deployment_id = aws_api_gateway_deployment.deployment.id
  rest_api_id   = aws_api_gateway_rest_api.minecraft.id
  stage_name    = "prod"
}

resource "aws_api_gateway_resource" "server" {
  parent_id   = aws_api_gateway_rest_api.minecraft.root_resource_id
  path_part   = "server"
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
}

module "servers_cors" {
  source      = "./api_cors_options"
  resource_id = aws_api_gateway_resource.servers.id
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
  cors_domain = aws_s3_bucket.site_bucket.bucket_domain_name
}

module "server_cors" {
  source      = "./api_cors_options"
  resource_id = aws_api_gateway_resource.server.id
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
  cors_domain = aws_s3_bucket.site_bucket.bucket_domain_name
}

module "server_start" {
  source       = "./api_resource"
  region       = data.aws_region.default.name
  account_id   = data.aws_caller_identity.current.account_id
  resource_id  = aws_api_gateway_resource.server.id
  rest_api_id  = aws_api_gateway_rest_api.minecraft.id
  dns_zone_id  = aws_route53_zone.domain.id
  dns_zone     = aws_route53_zone.domain.name
  subnets      = data.aws_subnets.subnets.ids
  cors_domain  = aws_s3_bucket.site_bucket.bucket_domain_name
  name         = "server_start"
  path         = "server"
  method       = "PUT"
  iam_actions = [
    "ec2:DescribeInstances",
    "ec2:RunInstances",
    "route53:ListResourceRecordSets",
    "route53:ChangeResourceRecordSets",
    "iam:PassRole", // "arn:aws:iam::647334721350:role/ssm"
    "ec2:CreateTags", // "arn:aws:ec2:eu-west-1:647334721350:instance/*"
  ]
}

module "server_stop" {
  source       = "./api_resource"
  region       = data.aws_region.default.name
  account_id   = data.aws_caller_identity.current.account_id
  resource_id  = aws_api_gateway_resource.server.id
  rest_api_id  = aws_api_gateway_rest_api.minecraft.id
  dns_zone_id  = aws_route53_zone.domain.id
  dns_zone     = aws_route53_zone.domain.name
  subnets      = data.aws_subnets.subnets.ids
  cors_domain  = aws_s3_bucket.site_bucket.bucket_domain_name
  name         = "server_stop"
  path         = "server"
  method       = "DELETE"
  iam_actions = [
    "ec2:DescribeInstances",
    "ec2:TerminateInstances",
    "route53:ListResourceRecordSets",
    "route53:ChangeResourceRecordSets",
    "route53:ListResourceRecordSets",
    "route53:ChangeResourceRecordSets",
  ]
}

module "server_status" {
  source       = "./api_resource"
  region       = data.aws_region.default.name
  account_id   = data.aws_caller_identity.current.account_id
  resource_id  = aws_api_gateway_resource.server.id
  rest_api_id  = aws_api_gateway_rest_api.minecraft.id
  dns_zone_id  = aws_route53_zone.domain.id
  dns_zone     = aws_route53_zone.domain.name
  subnets      = data.aws_subnets.subnets.ids
  cors_domain  = aws_s3_bucket.site_bucket.bucket_domain_name
  name         = "server_status"
  path         = "server"
  method       = "GET"
  iam_actions = [
    "ec2:DescribeInstances",
  ]
}

resource "aws_api_gateway_resource" "servers" {
  parent_id   = aws_api_gateway_rest_api.minecraft.root_resource_id
  path_part   = "servers"
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
}

module "servers_list" {
  source       = "./api_resource"
  region       = data.aws_region.default.name
  account_id   = data.aws_caller_identity.current.account_id
  resource_id  = aws_api_gateway_resource.servers.id
  rest_api_id  = aws_api_gateway_rest_api.minecraft.id
  dns_zone_id  = aws_route53_zone.domain.id
  dns_zone     = aws_route53_zone.domain.name
  subnets      = data.aws_subnets.subnets.ids
  cors_domain  = aws_s3_bucket.site_bucket.bucket_domain_name
  name         = "servers_list"
  path         = "servers"
  method       = "GET"
  iam_actions = [
    "elasticfilesystem:DescribeFileSystems",
    "ec2:DescribeInstances",
  ]
}
