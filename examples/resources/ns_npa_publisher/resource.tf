resource "ns_npa_publisher" "my_npapublisher" {
  lbrokerconnect                = false
  name                          = "Laverne Cummings"
  publisher_upgrade_profiles_id = 9
  tags = [
    {
      tag_id   = 4
      tag_name = "...my_tag_name..."
    },
  ]
}