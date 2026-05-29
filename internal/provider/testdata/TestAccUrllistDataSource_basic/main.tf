variable "name" {
  type = string
}

resource "netskope_urllist" "test" {
  name = var.name
  data = {
    urls = ["www.datasource-test.com"]
    type = "exact"
  }
}

data "netskope_urllist" "test" {
  id = netskope_urllist.test.id
}

data "netskope_urllist_list" "all" {
  depends_on = [netskope_urllist.test]
}
