resource "netskope_device_classification_steering_mapping" "my_deviceclassificationsteeringmapping" {
  onprem_detection_ids = [
    8
  ]
  steering_id = 10
}