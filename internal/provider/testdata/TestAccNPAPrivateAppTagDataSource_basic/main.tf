variable "tag_name" {
  type = string
}

resource "netskope_npa_private_app_tag" "test" {
  tag_name = var.tag_name
}

data "netskope_npa_private_app_tag" "test" {
  tag_id = netskope_npa_private_app_tag.test.tag_id
}
