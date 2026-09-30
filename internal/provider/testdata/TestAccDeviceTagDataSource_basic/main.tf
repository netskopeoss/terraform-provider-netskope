variable "name" {
  type = string
}

resource "netskope_device_tag" "test" {
  name        = var.name
  description = "Data source test tag"
}

data "netskope_device_tag" "test" {
  tag_id = netskope_device_tag.test.tag_id
}
