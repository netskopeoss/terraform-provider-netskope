variable "name" {
  type = string
}

resource "netskope_rbac_label" "test" {
  name  = var.name
  color = "#0294C9"
}
