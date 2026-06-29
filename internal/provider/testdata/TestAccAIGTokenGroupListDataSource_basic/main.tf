variable "name" {
  type = string
}

resource "netskope_aig_token_group" "test" {
  name        = var.name
  description = "Test token group"
}

data "netskope_aig_token_group_list" "test" {
  depends_on = [netskope_aig_token_group.test]
}
