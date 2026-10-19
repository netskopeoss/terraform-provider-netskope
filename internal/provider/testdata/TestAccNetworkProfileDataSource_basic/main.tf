variable "name" {
  type = string
}

resource "netskope_network_profile" "test" {
  name   = var.name
  values = ["192.168.1.0/24"]
}

data "netskope_network_profile" "test" {
  id = netskope_network_profile.test.id
}
