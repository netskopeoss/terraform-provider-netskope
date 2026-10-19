variable "name" {
  type = string
}

resource "netskope_network_profile" "test" {
  name        = var.name
  description = "Updated by acceptance test"
  values      = ["192.168.1.0/24", "10.10.0.0/16"]
}
