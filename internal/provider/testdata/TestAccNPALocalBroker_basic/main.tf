variable "name" {
  type = string
}

resource "netskope_npa_local_broker" "test" {
  local_broker_name = var.name
}
