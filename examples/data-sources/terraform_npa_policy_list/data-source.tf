data "terraform_npa_policy_list" "my_npapolicylist" {
  filter    = "...my_filter..."
  limit     = 5
  offset    = 2
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}