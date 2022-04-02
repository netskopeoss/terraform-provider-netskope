/*output "public_ip" {
  value = aws_instance.ec2_public.public_ip
}
*/

output "private_dns" {
  value = aws_instance.ec2_private.private_dns
}

output "npa_ip" {
  value = aws_instance.npa-publisher.public_ip
}