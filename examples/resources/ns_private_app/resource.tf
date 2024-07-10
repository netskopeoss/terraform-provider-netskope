resource "ns_private_app" "my_privateapp" {
  app_name             = "...my_app_name..."
  clientless_access    = true
  host                 = "...my_host..."
  private_app_id       = 3
  private_app_protocol = "...my_private_app_protocol..."
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
  real_host               = "...my_real_host..."
  trust_self_signed_certs = false
  use_publisher_dns       = false
}