variable "name" {
  type = string
}

resource "netskope_rbac_role" "test" {
  name        = var.name
  description = "Initial RBAC role"
  api_groups = [
    {
      api_group_id = 107
      permission   = "r"
    }
  ]
}
