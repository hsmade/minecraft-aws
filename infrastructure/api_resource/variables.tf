variable "account_id" {}
variable "region" {}
variable "resource_id" {}
variable "rest_api_id" {}
variable "name" {}
variable "path" {}
variable "method" {}
variable "cluster_name" {}
variable "dns_zone_id" {}
variable "dns_zone" {}
variable "subnets" {
  type = list(string)
}
variable "iam_actions" {
  type = list(string)
}
variable "cors_domain" {}
variable "ecs_sg_id" {}
