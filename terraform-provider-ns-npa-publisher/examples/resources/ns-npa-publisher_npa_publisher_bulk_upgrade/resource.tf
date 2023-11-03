resource "ns-npa-publisher_npa_publisher_bulk_upgrade" "my_npapublisherbulkupgrade" {
  publishers = {
    apply = {
      upgrade_request = false
    }
    id = [
      "...",
    ]
  }
}