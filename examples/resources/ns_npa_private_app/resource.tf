resource "ns_npa_private_app" "my_npaprivateapp" {
  allow_unauthenticated_cors = false
  allow_uri_bypass           = true
  app_name                   = "...my_app_name..."
  clientless_access          = true
  host                       = "...my_host..."
  is_user_portal_app         = true
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
  silent                  = "0"
  trust_self_signed_certs = true
  uribypass_header_value  = "...my_uribypass_header_value..."
  use_publisher_dns       = true
}