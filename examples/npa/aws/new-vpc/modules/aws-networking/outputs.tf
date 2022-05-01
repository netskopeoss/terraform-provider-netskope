output "vpc" {
  value = module.vpc
}

output "external_sg_id" {
  value = aws_security_group.external-sg.id
}

output "internal_sg_id" {
  value = aws_security_group.intenral-sg.id
}