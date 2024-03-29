resource "terraform_private_app_tag" "my_privateapptag" {
  id = "e9c5e434-7397-41ca-b5d5-462042ad5126"
  ids = [
    "...",
  ]
  publisher_tags = [
    {
      tag_id   = 8
      tag_name = "...my_tag_name..."
    },
  ]
  tags = [
    {
      tag_id   = 4
      tag_name = "...my_tag_name..."
    },
  ]
}