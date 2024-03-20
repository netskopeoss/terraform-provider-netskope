data "terraform_npa_policy_list" "my_npapolicylist" {
  filter    = "...my_filter..."
  limit     = 7
  offset    = 10
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}