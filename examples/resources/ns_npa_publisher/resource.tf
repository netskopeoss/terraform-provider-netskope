resource "ns_npa_publisher" "my_npapublisher" {
  lbrokerconnect                = true
  name                          = "Miss Arnold Runolfsdottir V"
  publisher_upgrade_profiles_id = 2
  tags = [
    {
      tag_id   = 9
      tag_name = "...my_tag_name..."
    },
  ]
}