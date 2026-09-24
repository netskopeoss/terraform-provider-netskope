variable "name" {
  type = string
}

resource "netskope_device_tag" "test" {
  name        = var.name
  description = "List data source test tag"
}

data "netskope_device_tag_list" "all" {
  depends_on = [netskope_device_tag.test]
}
