data "ns_policy_group_list" "my_policygrouplist" {
  filter    = "...my_filter..."
  limit     = 3
  offset    = 10
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}