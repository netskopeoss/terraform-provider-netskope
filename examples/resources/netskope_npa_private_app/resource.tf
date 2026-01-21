resource "netskope_npa_private_app" "my_npaprivateapp" {
  allow_unauthenticated_cors = false
  allow_uri_bypass           = false
  bypass_uris = [
    "..."
  ]
  clientless_access    = false
  is_user_portal_app   = true
  private_app_hostname = "...my_private_app_hostname..."
  private_app_name     = "...my_private_app_name..."
  private_app_protocol = "...my_private_app_protocol..."
  protocols = [
    {
      created_at = "2022-05-19 21:03:07.125000+00:00"
      id         = 1
      port       = "...my_port..."
      protocol   = "tcp"
      service_id = 9
      updated_at = "2022-05-19 21:03:07.125000+00:00"
    }
  ]
  publishers = [
    {
      publisher_id   = "...my_publisher_id..."
      publisher_name = "pub01.local"
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