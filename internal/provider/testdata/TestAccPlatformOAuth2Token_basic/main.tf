variable "client_id" {
  type = string
}

variable "client_secret" {
  type      = string
  sensitive = true
}

data "netskope_platform_oauth2_token" "test" {
  client_id     = var.client_id
  client_secret = var.client_secret
}
