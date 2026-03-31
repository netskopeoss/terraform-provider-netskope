variable "name" {
  type = string
}

variable "access_mode" {
  type = string
}

resource "netskope_npa_local_broker" "test" {
  local_broker_name    = var.name
  access_via_public_ip = var.access_mode
}
