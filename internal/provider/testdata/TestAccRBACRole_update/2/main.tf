variable "name" {
  type = string
}

resource "netskope_rbac_role" "test" {
  name        = var.name
  description = "Updated RBAC role"
  api_groups = [
    {
      api_group_id = 107
      permission   = "rw"
    },
    {
      api_group_id = 106
      permission   = "r"
    }
  ]
}
