resource "terraform_npa_publishers_bulk_upgrade" "my_npapublishersbulkupgrade" {
  publishers = {
    apply = {
      upgrade_request = true
    }
    id = [
      "...",
    ]
  }
}