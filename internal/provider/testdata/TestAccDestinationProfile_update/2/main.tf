variable "name" {
  type = string
}

resource "netskope_destination_profile" "test" {
  name        = var.name
  description = "Updated by acceptance test"
  type        = "insensitive"
  values      = ["example.com", "updated.example.com"]
}
