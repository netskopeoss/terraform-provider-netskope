resource "terraform_private_app_tag" "my_privateapptag" {
  id = "01266e09-8614-4b3f-a836-66602262d902"
  ids = [
    "...",
  ]
  publisher_tags = [
    {
      tag_id   = 2
      tag_name = "...my_tag_name..."
    },
  ]
  tags = [
    {
      tag_id   = 7
      tag_name = "...my_tag_name..."
    },
  ]
}