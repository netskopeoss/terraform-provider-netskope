resource "netskope_urllist" "my_urllist" {
  data = {
    type = "exact"
    urls = [
      "www.example.com"
    ]
  }
  name = "...my_name..."
}