resource "terraform_private_app" "my_privateapp" {
  app_name          = "...my_app_name..."
  clientless_access = true
  host              = "...my_host..."
  protocols = [
    {
      port = "...my_port..."
      type = "...my_type..."
    },
  ]
  publishers = [
    {
      publisher_id   = "...my_publisher_id..."
      publisher_name = "...my_publisher_name..."
    },
  ]
  publisher_tags = [
    {
      tag_name = "...my_tag_name..."
    },
  ]
  real_host               = "...my_real_host..."
  trust_self_signed_certs = true
  use_publisher_dns       = true
}