variable "home_ip" {
  type        = string
  description = "IP to allow access to"
}

variable "domain_name" {
  type        = string
  description = "domain name to create records one for the servers"
}

variable "whitelist" {
  type        = list(string)
  description = "uuids or accounts to whitelist"
}