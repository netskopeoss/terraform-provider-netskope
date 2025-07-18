resource "netskope_npa_publishers_bulk_profile_updates" "my_npapublishersbulkprofileupdates" {
  publishers = {
    apply = {
      publisher_upgrade_profiles_id = 1
    }
    publisher_id = [
      10
    ]
  }
}