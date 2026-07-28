variable "name" {
  type = string
}

resource "netskope_rbac_role" "test" {
  name        = var.name
  description = "Acceptance test RBAC role"
  api_groups = [
    {
      api_group_id = 107
      permission   = "r"
    }
  ]
}

data "netskope_rbac_role_list" "test" {
  depends_on = [netskope_rbac_role.test]
}
