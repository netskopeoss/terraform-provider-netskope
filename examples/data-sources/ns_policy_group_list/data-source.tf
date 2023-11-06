data "ns_policy_group_list" "my_policygrouplist" {
  filter    = "...my_filter..."
  limit     = 5
  offset    = 8
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}