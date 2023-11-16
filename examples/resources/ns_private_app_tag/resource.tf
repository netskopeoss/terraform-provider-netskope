resource "ns_private_app_tag" "my_privateapptag" {
  id = "6254595b-9e34-4d03-9d9e-11ab068c00e0"
  ids = [
    "...",
  ]
  publisher_tags = [
    {
      tag_id   = 3
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