resource "netskope_npa_policy_groups" "my_npapolicygroups" {
  group_name = "...my_group_name..."
  group_order = {
    group_order = {
      group_id = "1"
      order    = "before"
    }
  }
  modify_by   = "...my_modify_by..."
  modify_type = "...my_modify_type..."
  silent      = "1"
}