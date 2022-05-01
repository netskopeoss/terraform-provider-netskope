variable "publisher" {
  description = "Publisher Name"
  default     = "ns-publisher-tf"
  type        = string
}

variable "privateapp" {
  description = "Private Application Name"
  default     = "privateapp-tf"
  type        = string
}

variable "privateapp_host" {
  description = "Private Application Host"
  default     = "*.ec2.internal"
  type        = string
}

variable "privateapp_tcp" {
  description = "TCP Ports"
  default     = "22, 443" //Exam[ple Showing List
  type        = string
}

variable "privateapp_udp" {
  description = "UDP Ports"
  default     = "123-124" //Example Showing range
  type        = string
}

variable "aws_instance_type" {
  description = "AWS Instance Type"
  default     = "t3.medium" //Reccomended Publisher Type
  type        = string
}

variable "aws_key_name" {
  description = "AWS SSH Key Name"
  type        = string
}

variable "aws_subnet" {
  description = "AWS Subnet Id"
  type        = string
}

variable "aws_security_group" {
  description = "AWS SG Id"
  type        = string
}
