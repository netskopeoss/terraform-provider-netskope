data "netskope_npa_rules_list" "my_nparuleslist" {
  filter    = "...my_filter..."
  limit     = 0
  offset    = 5
  sortby    = "...my_sortby..."
  sortorder = "...my_sortorder..."
}