data "ns_policy_group_list" "my_policygrouplist" {
  filter    = "...my_filter..."
  limit     = 7
  offset    = 9
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}