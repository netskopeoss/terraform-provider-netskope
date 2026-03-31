variable "name" {
  type = string
}

resource "netskope_npa_publisher" "test" {
  publisher_name = var.name
}

data "netskope_npa_publisher" "test" {
  publisher_id = netskope_npa_publisher.test.publisher_id
}
