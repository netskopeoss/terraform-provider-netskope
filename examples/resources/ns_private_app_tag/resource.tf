resource "ns_private_app_tag" "my_privateapptag" {
  id = "595b9e34-d035-4d9e-91ab-068c00e045fc"
  ids = [
    "...",
  ]
  publisher_tags = [
    {
      tag_id   = 5
      tag_name = "...my_tag_name..."
    },
  ]
  tag_id = 2
  tags = [
    {
      tag_id   = 6
      tag_name = "...my_tag_name..."
    },
  ]
}