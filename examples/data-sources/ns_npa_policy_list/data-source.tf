data "ns_npa_policy_list" "my_npapolicylist" {
  filter    = "...my_filter..."
  limit     = 4
  offset    = 1
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}