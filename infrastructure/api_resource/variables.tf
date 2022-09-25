#variable "bucket" {}
variable "account_id" {}
variable "region" {}
variable "resource_id" {}
variable "rest_api_id" {}
variable "name" {}
variable "path" {}
variable "method" {}
variable "cluster_name" {}
variable "dns_zone_id" {}
variable "iam_actions" {
  type = list(string)
}