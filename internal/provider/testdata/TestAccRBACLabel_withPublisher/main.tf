variable "label_name" {
  type = string
}

variable "publisher_name" {
  type = string
}

resource "netskope_rbac_label" "test" {
  name  = var.label_name
  color = "#0294C9"
}

resource "netskope_npa_publisher" "test" {
  publisher_name = var.publisher_name
  label_ids      = [netskope_rbac_label.test.label_id]
}
