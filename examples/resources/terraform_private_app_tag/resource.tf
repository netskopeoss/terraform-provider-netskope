resource "terraform_private_app_tag" "my_privateapptag" {
  id = "ae755bc5-86df-4fb0-9c98-0090946a7d0e"
  ids = [
    "...",
  ]
  tags = [
    {
      tag_id   = 9
      tag_name = "...my_tag_name..."
    },
  ]
}