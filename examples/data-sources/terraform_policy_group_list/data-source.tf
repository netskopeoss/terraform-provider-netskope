data "terraform_policy_group_list" "my_policygrouplist" {
  filter    = "...my_filter..."
  limit     = 7
  offset    = 6
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}