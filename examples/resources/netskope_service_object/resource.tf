resource "netskope_service_object" "my_serviceobject" {
  description = "...my_description..."
  name        = "...my_name..."
  protocols = {
    icmp = true
    tcp = [
      "..."
    ]
    tcp_udp = [
      "..."
    ]
    udp = [
      "..."
    ]
  }
}