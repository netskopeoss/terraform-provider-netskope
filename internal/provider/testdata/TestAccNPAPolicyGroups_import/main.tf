variable "name" {
  type = string
}

resource "netskope_npa_policy_groups" "test" {
  group_name = var.name

  group_order = {
    group_id = "2"
    order    = "after"
  }
}
