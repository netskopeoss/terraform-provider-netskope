resource "netskope_npa_private_app" "my_npaprivateapp" {
  allow_unauthenticated_cors = false
  allow_uri_bypass           = false
  app_option = {
    # ...
  }
  bypass_uris = [
    "..."
  ]
  clientless_access  = false
  custom_host        = "...my_custom_host..."
  hide_app_in_portal = false
  is_user_portal_app = true
  paths = [
    "..."
  ]
  private_app_hostname = "...my_private_app_hostname..."
  private_app_name     = "...my_private_app_name..."
  private_app_protocol = "...my_private_app_protocol..."
  protocols = [
    {
      port     = "...my_port..."
      protocol = "tcp"
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
  trust_self_signed_certs   = false
  upgrade_insecure_requests = true
  uribypass_header_value    = "...my_uribypass_header_value..."
  use_publisher_dns         = true
}