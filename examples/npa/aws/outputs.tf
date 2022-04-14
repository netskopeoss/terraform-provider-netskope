output "intenral_ec2_connection_string" {
  description = "Copy/Paste/Enter - You are in the private ec2 instance"
  value       = "ssh -i ${module.ssh-key.key_name}.pem ec2-user@${module.ec2.private_dns}"
}

output "pub01_connection_string" {
  description = "Copy/Paste/Enter - You are in the Publisher01 npa instance"
  value       = "ssh -i ${module.ssh-key.key_name}.pem centos@${module.ec2.pub01_ip}"
}

output "pub02_connection_string" {
  description = "Copy/Paste/Enter - You are in the Publisher01 npa instance"
  value       = "ssh -i ${module.ssh-key.key_name}.pem centos@${module.ec2.pub02_ip}"
}

output "publisher01_details" {
  description = "Name and Token for the Publisher"
  value = {
    name  = "${module.publishers.pub01_name}"
    token = "${module.publishers.pub01_token}"
  }
}

output "publisher02_details" {
  description = "Name and Token for the Publisher"
  value = {
    name  = "${module.publishers.pub02_name}"
    token = "${module.publishers.pub02_token}"
  }
}