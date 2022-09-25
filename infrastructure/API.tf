module "server_start" {
  source           = "./api_resource"
  region           = data.aws_region.default.name
  account_id       = data.aws_caller_identity.current.account_id
  rest_api_id      = aws_api_gateway_rest_api.minecraft.id
  rest_api_root_id = aws_api_gateway_rest_api.minecraft.root_resource_id
  name = "start"
  path = "server"
  method = "PUT"
  iam_actions = ["s3:listBucket"]
}

module "server_stop" {
  source           = "./api_resource"
  region           = data.aws_region.default.name
  account_id       = data.aws_caller_identity.current.account_id
  rest_api_id      = aws_api_gateway_rest_api.minecraft.id
  rest_api_root_id = aws_api_gateway_rest_api.minecraft.root_resource_id
  name = "stop"
  path = "server"
  method = "PUT"
  iam_actions = ["s3:listBucket"]
}

module "server_status" {
  source           = "./api_resource"
  region           = data.aws_region.default.name
  account_id       = data.aws_caller_identity.current.account_id
  rest_api_id      = aws_api_gateway_rest_api.minecraft.id
  rest_api_root_id = aws_api_gateway_rest_api.minecraft.root_resource_id
  name = "status"
  path = "server"
  method = "GET"
  iam_actions = ["s3:listBucket"]
}

module "servers_list" {
  source           = "./api_resource"
  region           = data.aws_region.default.name
  account_id       = data.aws_caller_identity.current.account_id
  rest_api_id      = aws_api_gateway_rest_api.minecraft.id
  rest_api_root_id = aws_api_gateway_rest_api.minecraft.root_resource_id
  name = "list"
  path = "servers"
  method = "GET"
  iam_actions = ["s3:listBucket"]
}
