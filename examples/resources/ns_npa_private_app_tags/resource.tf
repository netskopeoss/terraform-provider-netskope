resource "ns_npa_private_app_tags" "my_npaprivateapptags" {
  tag_id = 5
  tags = [
    {
      tag_name           = "...my_tag_name..."
      x_speakeasy_entity = "{ \"see\": \"documentation\" }"
    }
  ]
  x_speakeasy_entity = "{ \"see\": \"documentation\" }"
}