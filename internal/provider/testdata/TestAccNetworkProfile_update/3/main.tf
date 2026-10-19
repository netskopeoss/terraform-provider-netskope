variable "name" {
  type = string
}

resource "netskope_network_profile" "test" {
  name   = var.name
  values = ["10.10.0.0/16"]
}
