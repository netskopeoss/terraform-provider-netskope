output "private_dns" {
  value = aws_instance.ec2_internal.private_dns
}

output "pub01_ip" {
  value = aws_instance.npa-publisher01.public_ip
}

output "pub02_ip" {
  value = aws_instance.npa-publisher01.public_ip
}