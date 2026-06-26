resource "netskope_aig_ai_provider" "my_aigaiprovider" {
  certificate = "-----BEGIN CERTIFICATE-----\nxxxxxxxx\n-----END CERTIFICATE-----"
  host        = "ai.backend.dev.company.com"
  name        = "cust-dev-ollama"
  port        = 443
  protocol    = "https-system"
  schema      = "openai"
}