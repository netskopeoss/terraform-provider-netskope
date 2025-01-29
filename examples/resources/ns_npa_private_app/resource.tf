resource "ns_npa_private_app" "my_npaprivateapp" {
  allow_unauthenticated_cors = false
  allow_uri_bypass           = false
  app_name                   = "Webserver"
  app_option = {
    # ...
  }
  bypass_uris = [
    "..."
  ]
  clientless_access    = false
  is_user_portal_app   = true
  private_app_hostname = "...my_private_app_hostname..."
  protocols = [
    {
      port     = 22
      protocol = "...my_protocol..."
    }
  ]
  publishers = [
    {
      publisher_id   = "...my_publisher_id..."
      publisher_name = "...my_publisher_name..."
    }
  ]
  real_host = "...my_real_host..."
  tags = [
    {
      tag_name = "...my_tag_name..."
    }
  ]
  trust_self_signed_certs = false
  uribypass_header_value  = "...my_uribypass_header_value..."
  use_publisher_dns       = true
}