variable "name" {
  type = string
}

resource "netskope_device_classification_tag" "test" {
  name = "${var.name}-tag"
}

resource "netskope_device_classification_rule" "test" {
  name  = var.name
  label = netskope_device_classification_tag.test.name
  os    = "mac"

  conditions = jsonencode({
    "$and" = [
      {
        "$and" = [
          {
            "$and" = [
              {
                "min_os_version_check" = {
                  "min_os_version" = "12.0.0"
                }
              }
            ]
          }
        ]
      }
    ]
  })
}
