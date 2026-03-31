variable "hostname" {
  type = string
}

resource "netskope_npa_local_broker_config" "test" {
  hostname = var.hostname
}
