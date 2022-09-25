#variable "bucket" {}
variable "account_id" {}
variable "region" {}
variable "rest_api_id" {}
variable "rest_api_root_id" {}
variable "name" {}
variable "path" {}
variable "method" {}
variable "iam_actions" {
  type = list(string)
}