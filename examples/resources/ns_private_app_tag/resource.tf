resource "ns_private_app_tag" "my_privateapptag" {
  id = "59625459-5b9e-434d-835d-9e11ab068c00"
  ids = [
    "...",
  ]
  publisher_tags = [
    {
      tag_id   = 10
      tag_name = "...my_tag_name..."
    },
  ]
  tags = [
    {
      tag_id   = 0
      tag_name = "...my_tag_name..."
    },
  ]
}