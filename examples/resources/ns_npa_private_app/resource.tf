resource "ns_npa_private_app" "my_npaprivateapp" {
  allow_unauthenticated_cors = false
  app_name                   = "...my_app_name..."
  clientless_access          = false
  is_user_portal_app         = false
  private_app_hostname       = "webserver.local"
  private_app_protocol       = "http"
  protocols = [
    {
      created_at = "2024-07-26T18:11:06.809Z"
      id         = 108
      port       = 80
      protocol   = "tcp"
      service_id = 67
      updated_at = "2024-07-26T18:11:06.809Z"
    },
  ]
  publishers = [
    {
      publisher_id   = 10
      publisher_name = "webservers"
    },
  ]
  real_host               = ""
  trust_self_signed_certs = true
  use_publisher_dns       = false
}