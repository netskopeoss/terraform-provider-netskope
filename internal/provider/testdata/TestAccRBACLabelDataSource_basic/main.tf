variable "name" {
  type = string
}

resource "netskope_rbac_label" "test" {
  name  = var.name
  color = "#0294C9"
}

data "netskope_rbac_label" "test" {
  label_id = netskope_rbac_label.test.label_id
}
