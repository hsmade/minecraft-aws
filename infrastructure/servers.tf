module "pim" {
  source            = "./server"
  name              = "pim"
  bucket_url        = "${data.aws_s3_bucket.minecraft.bucket_domain_name}/pim.tgz"
  setup_image       = "${aws_ecr_repository.scripts.repository_url}:latest"
  teardown_image    = "${aws_ecr_repository.scripts.repository_url}:latest"
  region            = data.aws_region.default.name
  bucket_name       = var.bucket
  account_id        = data.aws_caller_identity.current.account_id
  sidecars_role_arn = aws_iam_role.sidecars.arn
  task_role_arn     = aws_iam_role.task_role.arn
}
