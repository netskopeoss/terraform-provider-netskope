variable "name" {
  type = string
}

resource "netskope_destination_profile" "test" {
  name   = var.name
  type   = "insensitive"
  values = ["example.com"]
}
