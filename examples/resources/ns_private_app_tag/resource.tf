resource "ns_private_app_tag" "my_privateapptag" {
  id = "96254595-b9e3-44d0-b5d9-e11ab068c00e"
  ids = [
    "...",
  ]
  publisher_tags = [
    {
      tag_id   = 0
      tag_name = "...my_tag_name..."
    },
  ]
  tags = [
    {
      tag_id   = 3
      tag_name = "...my_tag_name..."
    },
  ]
}