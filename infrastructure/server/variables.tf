variable "name" {}
variable "backup_image" {
  default = "itzg/mc-backup:latest"
}
variable "setup_image" {
  default = "scripts"
}
variable "teardown_image" {
  default = "scripts"
}
variable "main_image" {
  default = "itzg/minecraft-server:latest"
}
variable "account_id" {}
variable "region" {}
variable "bucket_name" {}
variable "execution_role_arn" {}
variable "sidecars_role_arn" {}
variable "minecraft_version" {}
variable "forge_version" {}
variable "minecraft_type" {}