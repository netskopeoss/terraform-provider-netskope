variable "name" {
  type = string
}

resource "netskope_aig_mcp_server" "test" {
  name     = var.name
  host     = "mcp-backend.example.com"
  port     = 8080
  path     = "/mcp"
  protocol = "http"
}
