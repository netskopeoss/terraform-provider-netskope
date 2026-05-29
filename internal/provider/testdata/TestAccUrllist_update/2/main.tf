variable "name" {
  type = string
}

resource "netskope_urllist" "test" {
  name = "${var.name}-updated"
  data = {
    urls = ["www.example.com", "www.new.com", "www.added.com"]
    type = "exact"
  }
}
