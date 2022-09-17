module "pim" {
  source         = "./server"
  name           = "pim"
  bucket_url     = "${data.aws_s3_bucket.minecraft.bucket_domain_name}/pim.tgz"
  setup_image    = "${aws_ecr_repository.scripts.repository_url}/scripts:latest"
  teardown_image = "${aws_ecr_repository.scripts.repository_url}/scripts:latest"
}