variable "name" {}
variable "bucket_url" {}
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