module "storage_list" {
  source           = "./api_resource"
  region           = data.aws_region.default.name
  account_id       = data.aws_caller_identity.current.account_id
  bucket           = var.bucket
  rest_api_id      = aws_api_gateway_rest_api.minecraft.id
  rest_api_root_id = aws_api_gateway_rest_api.minecraft.root_resource_id
  name = "storage_list"
  path = "servers"
  method = "GET"
  iam_actions = ["s3:listBucket"]
}