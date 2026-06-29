resource "netskope_aig_mcp_server" "my_aigmcpserver" {
  certificate = "-----BEGIN CERTIFICATE-----\nxxxxxxxx\n-----END CERTIFICATE-----"
  host        = "api.githubcopilot.com"
  name        = "mcp-cust-github"
  path        = "/mcp"
  port        = 443
  prompts = [
    "AssignCodingAgent",
    "issue_to_fix_workflow",
  ]
  protocol = "https-system"
  resources = [
    "ui://send-message-input.html",
  ]
  schema = "mcp-http"
  tools = [
    "add_comment_to_pending_review",
    "add_issue_comment",
  ]
}