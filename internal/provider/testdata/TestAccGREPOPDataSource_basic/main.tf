variable "pop_id" {
  type = string
}

data "netskope_grepop" "test" {
  pop_id = var.pop_id
}
