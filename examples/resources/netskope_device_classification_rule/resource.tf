resource "netskope_device_classification_rule" "my_deviceclassificationrule" {
  conditions = "{ \"see\": \"documentation\" }"
  label      = "...my_label..."
  name       = "...my_name..."
  os         = "mac"
}