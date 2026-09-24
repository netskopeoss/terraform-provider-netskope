variable "name" {
  type = string
}

resource "netskope_device_tag" "test" {
  name        = var.name
  description = "Acceptance test device tag"
}
