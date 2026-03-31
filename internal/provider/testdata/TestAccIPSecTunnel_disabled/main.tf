variable "name" {
  type = string
}

variable "source_ip" {
  type = string
}

resource "netskope_ip_sec_tunnel" "test" {
  site            = var.name
  source_ip       = var.source_ip
  source_identity = "${var.name}.example.com"
  psk             = "TestPreSharedKey123!"
  encryption      = "AES128-CBC"
  pop_names       = ["lon1", "lon2"]
  enabled         = false
}
