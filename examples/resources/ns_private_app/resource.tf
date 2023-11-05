resource "ns_private_app" "my_privateapp" {
  app_name          = "...my_app_name..."
  clientless_access = false
  host              = "...my_host..."
  protocols = [
    {
      id         = 4
      port       = "...my_port..."
      service_id = 3
      transport  = "...my_transport..."
      type       = "...my_type..."
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
  real_host = "...my_real_host..."
  tags = [
    {
      tag_id   = 1
      tag_name = "...my_tag_name..."
    },
  ]
  trust_self_signed_certs = false
  use_publisher_dns       = true
}