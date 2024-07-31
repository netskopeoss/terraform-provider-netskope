resource "ns_npa_publisher" "my_npapublisher" {
  lbrokerconnect                = false
  publisher_name                = "npa_publisher_1"
  publisher_upgrade_profiles_id = 1
}