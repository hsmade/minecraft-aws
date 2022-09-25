resource "aws_api_gateway_resource" "server" {
  parent_id   = aws_api_gateway_rest_api.minecraft.root_resource_id
  path_part   = "server"
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
}

module "server_start" {
  source      = "./api_resource"
  region      = data.aws_region.default.name
  account_id  = data.aws_caller_identity.current.account_id
  resource_id = aws_api_gateway_resource.server.id
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
  name        = "start"
  path        = "server"
  method      = "PUT"
  iam_actions = ["s3:listBucket"]
}

module "server_stop" {
  source      = "./api_resource"
  region      = data.aws_region.default.name
  account_id  = data.aws_caller_identity.current.account_id
  resource_id = aws_api_gateway_resource.server.id
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
  name        = "stop"
  path        = "server"
  method      = "DELETE"
  iam_actions = ["s3:listBucket"]
}

module "server_status" {
  source      = "./api_resource"
  region      = data.aws_region.default.name
  account_id  = data.aws_caller_identity.current.account_id
  resource_id = aws_api_gateway_resource.server.id
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
  name        = "status"
  path        = "server"
  method      = "GET"
  iam_actions = ["s3:listBucket"]
}

resource "aws_api_gateway_resource" "servers" {
  parent_id   = aws_api_gateway_rest_api.minecraft.root_resource_id
  path_part   = "servers"
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
}

module "servers_list" {
  source      = "./api_resource"
  region      = data.aws_region.default.name
  account_id  = data.aws_caller_identity.current.account_id
  resource_id = aws_api_gateway_resource.servers.id
  rest_api_id = aws_api_gateway_rest_api.minecraft.id
  name        = "list"
  path        = "servers"
  method      = "GET"
  iam_actions = ["s3:listBucket"]
}
