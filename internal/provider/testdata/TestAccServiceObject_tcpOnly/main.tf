variable "name" {
  type = string
}

# Regression test for issue #119: omitted protocols must not be sent as empty
# arrays. Before the fix, udp=[] and tcp_udp=[] were sent to the API, which
# interpreted them as "Any port" and recorded them on the object.
resource "netskope_service_object" "test" {
  name        = var.name
  description = "Regression test issue 119 - tcp only"
  protocols = {
    tcp = ["443"]
  }
}
