# AWS VPC and Netskope Publisher Creation
This is an example plan that uses modules to create AWS infrastructure and with Netskope Publishers and a Private App that is used to reach to internal resources.

## Providers

- [Netskope Terraform Provider](https://github.com/netskopeoss/terraform-provider-netskope)


## Terraform Modules

Modules:

- publishers: Creates 2 Publisher objects in the Netskope Tenant
- ssh-key: Generates an ssh key pair to be used by EC2
- aws-network: Sets up a VPC that includes; 
    - IGW
    - NAT GWs
    - 2 external subnets 
    - 2 internal subnets
- ec2: Creates 2 NPA Publishers & 1 internal Linux Device
    - Uses the latest AMIs in the region 
    - Uses 2 external subnets and 1 internal subnet
   - NPA Publishers Registers themselves using the geberated tokens
   - ec2 in private subnet has egresses via NAT gateway
- privateapps: Creates a private app in the Netskope tenant 
    - Publishes the ec2 instance behind the created Publishers
    - Uses internal DNS name
    - Publishes tcp port 22

## Usage

1. Export Provider Environment Variables 
    ```sh 
    export NS_BaseURL=https://<tenant url>
    export NS_ApiToken=<apiv2 token>
    ```

1. Export Terraform Environment Variables  or create [.tfvars file](https://www.terraform.io/language/values/variables#variable-definitions-tfvars-files) (Optional)
    ```sh 
    export TF_VAR_project=<project name>
    export TF_VAR_REGION=<aws region>
    export AWS_PROFILE=<aws profile name>
    ```

1. Apply
    ```sh
    terraform init
    terraform apply
    ```
