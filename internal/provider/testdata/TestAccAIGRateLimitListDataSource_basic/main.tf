variable "name" {
  type = string
}

resource "netskope_aig_rate_limit" "test" {
  name = var.name

  criteria = {
    apply_on = "ai"
  }

  limit = {
    requests = 100
    unit     = "hour"
  }
}

data "netskope_aig_rate_limit_list" "test" {
  depends_on = [netskope_aig_rate_limit.test]
}
