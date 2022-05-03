variable "project" {
  description = "The project name; used to name objects."
  default     = "ns-tf-demo"
  type        = string
}

variable "region" {
  description = "AWS region"
  default     = "us-east-1"
  type        = string
}

//Optional 
/*
variable "profile" {
  description = "AWS Profile to use"
  default     = "Default"
  type        = string
}
*/