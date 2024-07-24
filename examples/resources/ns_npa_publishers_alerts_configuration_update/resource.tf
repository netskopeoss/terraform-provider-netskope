resource "ns_npa_publishers_alerts_configuration_update" "my_npapublishersalertsconfigurationupdate" {
  admin_users = [
    "...",
  ]
  event_types = [
    "UPGRADE_SUCCEEDED",
  ]
  selected_users = "abc@xyz.com,def@xyz.com"
}