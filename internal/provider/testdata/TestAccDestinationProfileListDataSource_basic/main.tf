variable "name" {
  type = string
}

resource "netskope_destination_profile" "test" {
  name   = var.name
  type   = "insensitive"
  values = ["example.com"]
}

data "netskope_destination_profile_list" "test" {
  depends_on = [netskope_destination_profile.test]
}
