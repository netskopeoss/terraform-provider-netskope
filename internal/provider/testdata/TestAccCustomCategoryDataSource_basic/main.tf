variable "name" {
  type = string
}

resource "netskope_custom_category" "test" {
  name        = var.name
  description = "Acceptance test custom category"
}

data "netskope_custom_category" "test" {
  id = netskope_custom_category.test.id
}
