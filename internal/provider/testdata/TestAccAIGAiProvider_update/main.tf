variable "name" {
  type = string
}

variable "host" {
  type = string
}

resource "netskope_aig_ai_provider" "test" {
  name     = var.name
  host     = var.host
  port     = 443
  protocol = "https-skip"
  schema   = "openai"
}
