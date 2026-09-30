variable "tag_name" {
  type = string
}

resource "netskope_npa_private_app_tag" "test" {
  tag_name = var.tag_name
}
