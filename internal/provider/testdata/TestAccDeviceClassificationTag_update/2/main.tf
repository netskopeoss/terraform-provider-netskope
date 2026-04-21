variable "name" {
  type = string
}

resource "netskope_device_classification_tag" "test" {
  name        = "${var.name}-updated"
  description = "Updated description"
}
