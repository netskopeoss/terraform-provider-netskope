// Filter AWS LInux AMIs for the latests
data "aws_ami" "amazon-linux-2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm*"]
  }
}

// Configure the EC2 instance in a internal subnet
resource "aws_instance" "ec2_internal" {
  ami                         = data.aws_ami.amazon-linux-2.id
  associate_public_ip_address = false
  instance_type               = "t2.micro"
  key_name                    = var.key_name
  subnet_id                   = var.vpc.private_subnets[1]
  vpc_security_group_ids      = [var.internal_sg_id]
  tags = {
    "Name" = "${var.project}-ec2-internal"
  }

}

// Filter Netskope Publishers AMIs for the latests
data "aws_ami" "npa-publisher" {
  most_recent = true
  owners      = ["679593333241"]

  filter {
    name   = "name"
    values = ["Netskope Private Access Publisher*"]
  }
}

// Create 1st Publisher
resource "aws_instance" "npa-publisher01" {
  ami                         = data.aws_ami.npa-publisher.id
  associate_public_ip_address = true
  instance_type               = "t2.micro"
  key_name                    = var.key_name
  subnet_id                   = var.vpc.public_subnets[0]
  vpc_security_group_ids      = [var.external_sg_id]
  user_data                   = var.token01


  tags = {
    "Name" = "${var.project}-npa-publisher01"
  }

}

// Create 2nd Publisher
resource "aws_instance" "npa-publisher02" {
  ami                         = data.aws_ami.npa-publisher.id
  associate_public_ip_address = true
  instance_type               = "t2.micro"
  key_name                    = var.key_name
  subnet_id                   = var.vpc.public_subnets[1]
  vpc_security_group_ids      = [var.external_sg_id]
  user_data                   = var.token02


  tags = {
    "Name" = "${var.project}-npa-publisher02"
  }

}
