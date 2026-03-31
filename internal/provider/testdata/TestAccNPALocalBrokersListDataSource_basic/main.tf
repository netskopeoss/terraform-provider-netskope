variable "name" {
  type = string
}

resource "netskope_npa_local_broker" "test" {
  local_broker_name = var.name
}

data "netskope_npa_local_brokers_list" "test" {
  depends_on = [netskope_npa_local_broker.test]
}
