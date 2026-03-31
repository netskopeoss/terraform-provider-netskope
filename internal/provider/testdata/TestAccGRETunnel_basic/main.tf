variable "name" {
  type = string
}

variable "source_ip" {
  type = string
}

resource "netskope_gre_tunnel" "test" {
  site      = var.name
  source_ip = var.source_ip
  pop_names = ["lon1", "lon2"]
}
