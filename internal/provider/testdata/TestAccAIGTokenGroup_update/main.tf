variable "name" {
  type = string
}

variable "description" {
  type = string
}

resource "netskope_aig_token_group" "test" {
  name        = var.name
  description = var.description
}
