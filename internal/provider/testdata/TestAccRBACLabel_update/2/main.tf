variable "name" {
  type = string
}

resource "netskope_rbac_label" "test" {
  name  = var.name
  color = "#FF5733"
}
