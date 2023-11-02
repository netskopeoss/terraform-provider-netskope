resource "npa-publisher_npa_publisher" "my_npapublisher" {
  lbrokerconnect                = true
  name                          = "Mr. Bessie Hickle DVM"
  publisher_upgrade_profiles_id = 5
  tags = [
    {
      tag_id   = 9
      tag_name = "...my_tag_name..."
    },
  ]
}