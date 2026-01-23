resource "netskope_npa_private_app_public_host" "my_npaprivateapppublichost" {
  clientless_access  = false
  host               = "...my_host..."
  is_user_portal_app = true
  protocols = [
    {
      port     = "...my_port..."
      protocol = "tcp"
    }
  ]
  real_host = "...my_real_host..."
}