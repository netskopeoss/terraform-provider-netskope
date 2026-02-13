resource "netskope_npa_publishers_alerts_configuration" "my_npapublishersalertsconfiguration" {
  admin_users = [
    "admin1@abc.com",
    "admin2@abc.com",
  ]
  event_types = [
    "CONNECTION_FAILED",
    "UPGRADE_STARTED",
  ]
  selected_users = "abc@xyz.com,def@xyz.com"
}