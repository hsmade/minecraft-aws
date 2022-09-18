module "pim" {
  source             = "./server"
  name               = "pim"
  setup_image        = "${aws_ecr_repository.scripts.repository_url}:latest"
  teardown_image     = "${aws_ecr_repository.scripts.repository_url}:latest"
  region             = data.aws_region.default.name
  bucket_name        = var.bucket
  account_id         = data.aws_caller_identity.current.account_id
  sidecars_role_arn  = aws_iam_role.ecs_sidecars.arn
  execution_role_arn = aws_iam_role.ecs_execution_role.arn
}
