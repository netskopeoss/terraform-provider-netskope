data "ns_policy_group_list" "my_policygrouplist" {
  filter    = "...my_filter..."
  limit     = 1
  offset    = 2
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}