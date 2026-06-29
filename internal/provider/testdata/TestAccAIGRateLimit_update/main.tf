variable "name" {
  type = string
}

variable "requests" {
  type = number
}

resource "netskope_aig_rate_limit" "test" {
  name = var.name

  criteria = {
    apply_on = "ai"
  }

  limit = {
    requests = var.requests
    unit     = "hour"
  }
}
