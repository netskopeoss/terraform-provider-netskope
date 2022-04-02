# AWS VPC and Netskope Publisher Creation
This is an example plan that uses modules to create AWS infrastructure and with a Netskope Publisher and Private Apps used to reach to private resources.

## Terraform Modules

Modules:

- ssh-key: Generates an ssh key pair
- publishers: Creates a Publisher object in the Netskope Tenant
- network: Sets up a VPC with IGWs, NAT GWs, public and private subnets and SG
- ec2: Creates an NPA Publisher using the latest AMI in a public subnet and an ec2 instance in a private subnet
- ec2 in private subnet has outgoing network access though the NAT gateway
- privateapps: Create a private app in the Netskope tenant publishing the ec2 instance behind the created Publisher

No requirements.

## Providers

Netskope custom provider 
