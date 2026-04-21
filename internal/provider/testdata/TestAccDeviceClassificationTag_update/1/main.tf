variable "name" {
  type = string
}

resource "netskope_device_classification_tag" "test" {
  name        = var.name
  description = "Initial description"
}
