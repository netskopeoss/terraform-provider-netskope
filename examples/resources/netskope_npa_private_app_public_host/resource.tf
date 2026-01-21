resource "netskope_npa_private_app_public_host" "my_npaprivateapppublichost" {
  clientless_access  = false
  host               = "...my_host..."
  is_user_portal_app = true
  protocols = [
    {
      created_at = "2022-05-19 21:03:07.125000+00:00"
      id         = 4
      port       = "...my_port..."
      protocol   = "tcp"
      service_id = 9
      updated_at = "2022-05-19 21:03:07.125000+00:00"
    }
  ]
  real_host = "...my_real_host..."
}