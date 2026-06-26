resource "netskope_aig_appliance" "my_aigappliance" {
  ai_provider_ids = [
    "54fa6af9-6862-4b99-99c2-e1e915661a10"
  ]
  host = "my-aig-appliance.example.com"
  mcp_server_ids = [
    "ea85edbd-25cc-4b14-a12e-52a9bdf7b7a4"
  ]
  name = "my-aig-appliance"
  ports = {
    http = {
      enable = false
      port   = 443
    }
    https = {
      enable = true
      port   = 443
    }
  }
  sku_addons = [
    {
      product_code = "NK-A-AIGW-10K"
      quantity     = 5
    }
  ]
}