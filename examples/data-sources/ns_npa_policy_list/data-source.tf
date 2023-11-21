data "ns_npa_policy_list" "my_npapolicylist" {
  filter    = "...my_filter..."
  limit     = 8
  offset    = 5
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}