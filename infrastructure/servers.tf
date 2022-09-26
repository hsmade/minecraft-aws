module "pim" {
  source             = "./server"
  name               = "pim"
  region             = data.aws_region.default.name
  bucket_name        = var.bucket
  account_id         = data.aws_caller_identity.current.account_id
  sidecars_role_arn  = aws_iam_role.ecs_sidecars.arn
  execution_role_arn = aws_iam_role.ecs_execution_role.arn
  minecraft_type     = "FORGE"
  minecraft_version  = "1.18.1"
  forge_version      = "39.0.59"
  subnets            = data.aws_subnets.subnets.ids
  whitelist          = var.whitelist
  tags = {
    "mod1": "1.2.3",
    "mod2": "2.3.4",
  }
}

module "test" {
  source             = "./server"
  name               = "test"
  region             = data.aws_region.default.name
  bucket_name        = var.bucket
  account_id         = data.aws_caller_identity.current.account_id
  sidecars_role_arn  = aws_iam_role.ecs_sidecars.arn
  execution_role_arn = aws_iam_role.ecs_execution_role.arn
  minecraft_type     = "FORGE"
  minecraft_version  = "1.18.1"
  forge_version      = "39.0.59"
  subnets            = data.aws_subnets.subnets.ids
  whitelist          = var.whitelist
}
