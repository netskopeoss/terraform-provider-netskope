variable "name" {
  type = string
}

resource "netskope_custom_category" "test" {
  name = var.name
}

data "netskope_custom_category_list" "test" {
  depends_on = [netskope_custom_category.test]
}
