data "external" "git_checkout" {
  program = ["${path.module}/get_sha.sh"]
}

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

module "server_start" {
  source       = "./api_resource"
  region       = data.aws_region.default.name
  account_id   = data.aws_caller_identity.current.account_id
  resource_id  = aws_api_gateway_resource.server.id
  rest_api_id  = aws_api_gateway_rest_api.minecraft.id
  cluster_name = aws_ecs_cluster.minecraft.name
  dns_zone_id  = aws_route53_zone.domain.id
  name         = "server_start"
  path         = "server"
  method       = "PUT"
  iam_actions = [
    "ecs:CreateTaskSet",
    "ecs:ListTasks",
    "ecs:DescribeTasks",
    "route53:ListResourceRecordSets",
    "route53:ChangeResourceRecordSets",
  ]
}

module "server_stop" {
  source       = "./api_resource"
  region       = data.aws_region.default.name
  account_id   = data.aws_caller_identity.current.account_id
  resource_id  = aws_api_gateway_resource.server.id
  rest_api_id  = aws_api_gateway_rest_api.minecraft.id
  cluster_name = aws_ecs_cluster.minecraft.name
  dns_zone_id  = aws_route53_zone.domain.id
  name         = "server_stop"
  path         = "server"
  method       = "DELETE"
  iam_actions = [
    "ecs:StopTask",
    "ecs:ListTasks",
    "ecs:DescribeTasks",
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
  cluster_name = aws_ecs_cluster.minecraft.name
  dns_zone_id  = aws_route53_zone.domain.id
  name         = "server_status"
  path         = "server"
  method       = "GET"
  iam_actions = [
    "ecs:ListTasks",
    "ecs:DescribeTasks",
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
  cluster_name = aws_ecs_cluster.minecraft.name
  dns_zone_id  = aws_route53_zone.domain.id
  name         = "servers_list"
  path         = "servers"
  method       = "GET"
  iam_actions = [
    "ecs:ListTaskDefinitionFamilies",
    "ecs:ListTasks",
    "ecs:DescribeTasks",
  ]
}