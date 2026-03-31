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

  options = {
    xff = {
      xff_enabled = true
      xff_ip_list = ["10.0.0.1", "10.0.0.2"]
    }
  }
}
