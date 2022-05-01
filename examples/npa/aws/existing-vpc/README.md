# Netskope Publisher and Private App Creation
This is an example plan that uses the Netskope Provider and the AWS Provider to deploy a Publisher in an existing VPC and configure a Private App inside a Netskope Tenant to use the new Publisher.

## Providers

- [Netskope Terraform Provider](https://github.com/netskopeoss/terraform-provider-netskope)


## Usage 
1. Open a CLI shell and navigate to the provider directory
    ```sh
    cd <path>/terraform-provider-netskope
    ```
1. Export Provider Env Variables 
    ```sh 
    export NS_BaseURL=https://<tenant url>
    export NS_ApiToken=<apiv2 token>
    ```
1. Pass required Variables to Terraform with Env Variables in a .tfvars file.
    ```sh
    aws_key_name = "<existing key name>"
    aws_subnet = "<existing subnet id>"
    aws_security_group = "<existing security group>"
    ```
1. Run Simple Plan
    ```sh
    cd examples/npa/aws/existing-vpc    
    terraform init
    terraform apply
    ```
1. Destroy Simple Objects
    ```sh
    terraform apply -destroy
    ```