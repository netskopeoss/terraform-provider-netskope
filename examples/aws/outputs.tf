/*
output "public_connection_string" {
  description = "Copy/Paste/Enter - You are in the matrix"
  value       = "ssh -i ${module.ssh-key.key_name}.pem ec2-user@${module.ec2.public_ip}"
}
*/

output "private_connection_string" {
  description = "Copy/Paste/Enter - You are in the private ec2 instance"
  value       = "ssh -i ${module.ssh-key.key_name}.pem ec2-user@${module.ec2.private_dns}"
}

output "npa_connection_string" {
  description = "Copy/Paste/Enter - You are in the private npa instance"
  value       = "ssh -i ${module.ssh-key.key_name}.pem centos@${module.ec2.npa_ip}"
}

output "publisher_details" {
    description = "Name and Token for the Publisher"
    value = {
    name         = "${module.publishers.pub_name}"
    token         = "${module.publishers.pub_token}"
  }  
}