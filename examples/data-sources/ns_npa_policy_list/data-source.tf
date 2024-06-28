data "ns_npa_policy_list" "my_npapolicylist" {
  filter    = "...my_filter..."
  limit     = 7
  offset    = 8
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}