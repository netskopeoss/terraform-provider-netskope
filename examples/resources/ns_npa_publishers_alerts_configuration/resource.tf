resource "ns_npa_publishers_alerts_configuration" "my_npapublishersalertsconfiguration" {
  admin_users = [
    "admin1@abc.com",
  ]
  event_types = [
    "UPGRADE_WILL_START",
  ]
  selected_users = "abc@xyz.com,def@xyz.com"
}