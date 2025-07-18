data "netskope_npa_policy_groups_list" "my_npapolicygroupslist" {
  fields    = "...my_fields..."
  filter    = "...my_filter..."
  limit     = 7
  offset    = 6
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}