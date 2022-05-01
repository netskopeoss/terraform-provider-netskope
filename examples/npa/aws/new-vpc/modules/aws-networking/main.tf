//Get AZs from Data Provider
data "aws_availability_zones" "available" {}

module "vpc" {
  source          = "terraform-aws-modules/vpc/aws"
  name            = "${var.project}-vpc"
  cidr            = "10.0.0.0/16"
  azs             = data.aws_availability_zones.available.names
  private_subnets = ["10.0.10.0/24", "10.0.20.0/24"]
  public_subnets  = ["10.0.0.0/25", "10.0.0.128/25"]
  //create_database_subnet_group = true
  enable_nat_gateway           = true
  single_nat_gateway           = true
  enable_dns_hostnames         = true
  enable_dns_support           = true
}

// External SG
resource "aws_security_group" "external-sg" {
  name        = "${var.project}-external-sg"
  description = "SSH inbound"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "Allow SSH 0.0.0.0/0"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project}-external-sg"
  }
}

// Internal SG
resource "aws_security_group" "intenral-sg" {
  name        = "${var.project}-intenral-sg"
  description = "Allow SSH inbound traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "Allow SSH 10.0.0.0/16"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project}-internal-sg"
  }
}