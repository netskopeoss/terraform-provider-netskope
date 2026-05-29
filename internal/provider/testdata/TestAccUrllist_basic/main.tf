variable "name" {
  type = string
}

resource "netskope_urllist" "test" {
  name = var.name
  data = {
    urls = ["www.example.com", "www.test.com"]
    type = "exact"
  }
}
