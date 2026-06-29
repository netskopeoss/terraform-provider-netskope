variable "name" {
  type = string
}

resource "netskope_aig_ai_provider" "test" {
  name     = var.name
  host     = "ai-backend.example.com"
  port     = 443
  protocol = "https-skip"
  schema   = "openai"
}

data "netskope_aig_ai_provider" "test" {
  id = netskope_aig_ai_provider.test.id
}
