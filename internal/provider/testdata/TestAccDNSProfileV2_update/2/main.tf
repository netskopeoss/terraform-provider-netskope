variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name        = var.name
  description = "Updated by acceptance test"
  log_traffic = "All DNS"
}
