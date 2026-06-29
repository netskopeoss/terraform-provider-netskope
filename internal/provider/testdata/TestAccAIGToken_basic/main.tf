variable "name" {
  type = string
}

resource "netskope_aig_token_group" "test" {
  name = var.name
}

resource "netskope_aig_token" "test" {
  name           = var.name
  token_group_id = netskope_aig_token_group.test.id
  expire_in = {
    unit  = "month"
    value = 1
  }
}
