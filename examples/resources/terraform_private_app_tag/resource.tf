resource "terraform_private_app_tag" "my_privateapptag" {
  id = "1ba65a2c-5874-40b5-86d4-8f6603f14abb"
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
      tag_id   = 8
      tag_name = "...my_tag_name..."
    },
  ]
}