variable "parent_name" {
  type = string
}

variable "child_name" {
  type = string
}

resource "netskope_rbac_label" "parent" {
  name  = var.parent_name
  color = "#0294C9"
}

resource "netskope_rbac_label" "child" {
  name      = var.child_name
  parent_id = netskope_rbac_label.parent.label_id
  color     = "#FF5733"
}
