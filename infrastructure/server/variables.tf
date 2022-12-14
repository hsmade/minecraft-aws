variable "name" {}
variable "backup_image" {
  default = "itzg/mc-backup:latest"
}
variable "main_image" {
  default = "itzg/minecraft-server:latest"
}
variable "rcon_image" {
  default = "itzg/rcon:latest"
}
variable "account_id" {}
variable "region" {}
variable "bucket_name" {}
variable "execution_role_arn" {}
variable "sidecars_role_arn" {}
variable "minecraft_version" {}
variable "forge_version" {}
variable "minecraft_type" {}
variable "subnets" {
  type = list(string)
}
variable "whitelist" {
  type = list(string)
}
variable "ops" {
  type = list(string)
}
variable "tags" {
  type    = map(string)
  default = {}
}
variable "efs_sg_id" {}
