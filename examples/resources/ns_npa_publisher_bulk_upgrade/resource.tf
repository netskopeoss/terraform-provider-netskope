resource "ns_npa_publisher_bulk_upgrade" "my_npapublisherbulkupgrade" {
  publishers = {
    apply = {
      upgrade_request = true
    }
    id = [
      "...",
    ]
  }
}