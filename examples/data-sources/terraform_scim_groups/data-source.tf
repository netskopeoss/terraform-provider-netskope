data "terraform_scim_groups" "my_scimgroups" {
  count       = 3
  filter      = "...my_filter..."
  start_index = 2
}