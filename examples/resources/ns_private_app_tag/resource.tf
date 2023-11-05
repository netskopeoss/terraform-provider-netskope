resource "ns_private_app_tag" "my_privateapptag" {
  id = "d9876c8f-263c-4e6c-97d4-b3a292affc65"
  ids = [
    "...",
  ]
  publisher_tags = [
    {
      tag_id   = 9
      tag_name = "...my_tag_name..."
    },
  ]
  tag_id = 7
  tags = [
    {
      tag_id   = 6
      tag_name = "...my_tag_name..."
    },
  ]
}