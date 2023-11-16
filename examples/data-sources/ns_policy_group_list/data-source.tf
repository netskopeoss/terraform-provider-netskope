data "ns_policy_group_list" "my_policygrouplist" {
  filter    = "...my_filter..."
  limit     = 9
  offset    = 7
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}