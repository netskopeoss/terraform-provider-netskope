resource "netskope_aig_rate_limit" "my_aigratelimit" {
  appliance_ids = [
    "0197d486-93b1-7502-86a2-fddd00cd92ab",
  ]
  criteria = {
    ai_provider_ids = [
      "0197d486-93b1-7502-86a2-fddd00cd92cd",
    ]
    apply_on = "ai"
    mcp_server_ids = [
      "0197d486-93b1-7502-86a2-fddd00cd92ff",
    ]
    models = [
      {
        type  = "exact"
        value = "gpt-3.5-turbo"
      }
    ]
    prompts = [
      {
        type  = "exact"
        value = "gpt-3.5-turbo"
      }
    ]
    resources = [
      {
        type  = "exact"
        value = "gpt-3.5-turbo"
      }
    ]
    token_group_ids = [
      "0197d486-93b1-7502-86a2-fddd00cd92ef",
    ]
    tools = [
      {
        type  = "exact"
        value = "gpt-3.5-turbo"
      }
    ]
  }
  limit = {
    requests = 100
    unit     = "hour"
  }
  name     = "corp-gpt-limit"
  response = "Rate limit exceeded. Try again later."
}