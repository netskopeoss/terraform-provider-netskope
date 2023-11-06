resource "ns_npa_publisher" "my_npapublisher" {
  lbrokerconnect                = false
  name                          = "Mrs. Cory Jakubowski"
  publisher_upgrade_profiles_id = 5
  tags = [
    {
      tag_id   = 6
      tag_name = "...my_tag_name..."
    },
  ]
}