resource "netskope_custom_category" "my_customcategory" {
  description = "...my_description..."
  excluded_destination_profiles = [
    "..."
  ]
  excluded_url_lists = [
    "..."
  ]
  included_destination_profiles = [
    "..."
  ]
  included_predefined_categories = [
    "..."
  ]
  included_url_lists = [
    "..."
  ]
  name = "...my_name..."
}