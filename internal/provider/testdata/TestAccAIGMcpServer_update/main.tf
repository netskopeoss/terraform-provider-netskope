variable "name" {
  type = string
}

variable "path" {
  type = string
}

resource "netskope_aig_mcp_server" "test" {
  name     = var.name
  host     = "mcp-backend.example.com"
  port     = 8080
  path     = var.path
  protocol = "http"
}
