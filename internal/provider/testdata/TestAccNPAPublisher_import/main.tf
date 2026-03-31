variable "name" {
  type = string
}

resource "netskope_npa_publisher" "test" {
  publisher_name = var.name
}
