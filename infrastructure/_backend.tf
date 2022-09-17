terraform {
  backend "s3" {
    key     = "terraform"
    encrypt = true
  }
}