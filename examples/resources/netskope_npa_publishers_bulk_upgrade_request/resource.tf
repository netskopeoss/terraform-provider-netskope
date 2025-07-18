resource "netskope_npa_publishers_bulk_upgrade_request" "my_npapublishersbulkupgraderequest" {
  publishers = {
    apply = {
      upgrade_request = true
    }
    publisher_id = [
      "1"
    ]
  }
}