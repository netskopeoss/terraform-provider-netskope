variable "name" {
  type = string
}

# Device Tag — manual tag that external tools assign to devices.
# Note: not deleted on terraform destroy (API restriction).
resource "netskope_device_tag" "test" {
  name        = var.name
  description = "Acceptance test device tag check condition"
}

resource "netskope_device_classification_tag" "test" {
  name = "${var.name}-class"
}

# Classification rule using device_tag_check — the condition pattern used in
# the device-tags-for-classification example.
resource "netskope_device_classification_rule" "test" {
  name  = var.name
  label = netskope_device_classification_tag.test.name
  os    = "windows"

  conditions = jsonencode({
    "$and" = [
      {
        "$and" = [
          {
            "$and" = [
              {
                "device_tag_check" = {
                  "tag_id" = netskope_device_tag.test.tag_id
                }
              }
            ]
          }
        ]
      }
    ]
  })
}
