resource "terraform_npa_publisher_upgrade_profile" "my_npapublisherupgradeprofile" {
  docker_tag         = "...my_docker_tag..."
  enabled            = true
  frequency          = "...my_frequency..."
  name               = "Mrs. Rosa Padberg"
  release_type       = "...my_release_type..."
  required           = "{ \"see\": \"documentation\" }"
  timezone           = "...my_timezone..."
  upgrade_profile_id = 10
}