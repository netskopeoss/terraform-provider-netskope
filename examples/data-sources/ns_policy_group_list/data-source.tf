data "ns_policy_group_list" "my_policygrouplist" {
  filter    = "...my_filter..."
  limit     = 8
  offset    = 5
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}