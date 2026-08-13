variable "name" {
  type = string
}

resource "netskope_device_classification_on_prem_detection" "test" {
  name = var.name

  config = jsonencode({
    "onpremcheck" = {
      "onprem_use_dns"                       = "1"
      "onprem_host"                          = "internal.example.com"
      "onprem_ip"                            = "10.0.0.1"
      "onprem_http_host"                     = ""
      "onprem_http_tcp_connection_timeout"   = "10"
      "onprem_additional_ips"                = []
      "onprem_additional_http_hosts"         = []
      "onprem_detection_method"              = "1"
      "onprem_egress_ips"                    = []
    }
  })
}

data "netskope_device_classification_on_prem_detection" "test" {
  onprem_id = netskope_device_classification_on_prem_detection.test.onprem_id
}
