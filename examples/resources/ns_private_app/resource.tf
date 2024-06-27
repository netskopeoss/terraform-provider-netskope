resource "ns_private_app" "my_privateapp" {
  app_name          = "...my_app_name..."
  clientless_access = false
  host              = "...my_host..."
  private_app_id    = 2
  protocols = [
    {
      port = "...my_port..."
      type = "...my_type..."
    },
  ]
  publisher_tags = [
    {
      tag_name = "...my_tag_name..."
    },
  ]
  publishers = [
    {
      publisher_id   = "...my_publisher_id..."
      publisher_name = "...my_publisher_name..."
    },
  ]
  real_host               = "...my_real_host..."
  trust_self_signed_certs = true
  use_publisher_dns       = true
}