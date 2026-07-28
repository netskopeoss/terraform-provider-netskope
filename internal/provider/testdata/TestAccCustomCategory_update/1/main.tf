variable "name" {
  type = string
}

resource "netskope_custom_category" "test" {
  name        = var.name
  description = "Initial description"
}
