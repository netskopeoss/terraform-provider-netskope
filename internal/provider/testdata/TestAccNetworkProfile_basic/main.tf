variable "name" {
  type = string
}

resource "netskope_network_profile" "test" {
  name        = var.name
  description = "Acceptance test network profile"
  values      = ["192.168.1.0/24", "10.0.0.1-10.0.0.254"]
}
