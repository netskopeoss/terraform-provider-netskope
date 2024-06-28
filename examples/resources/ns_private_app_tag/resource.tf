resource "ns_private_app_tag" "my_privateapptag" {
  ids = [
    "...",
  ]
  tag_id = 3
  tags = [
    {
      tag_name = "...my_tag_name..."
    },
  ]
}