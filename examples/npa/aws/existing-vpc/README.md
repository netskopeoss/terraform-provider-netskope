# Netskope Publisher and Private App Creation - Existing VPC
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
1. Export Required Terraform Environment Variables  or create [terraform.tfvars file](https://www.terraform.io/language/values/variables#variable-definitions-tfvars-files)
    ```sh
    export TF_VAR_aws_key_name = "<existing ssh key name>"
    export TF_VAR_aws_subnet = "<existing subnet id>"
    export TF_VAR_aws_security_group = "<existing security group>"
    ```
1. Include Optional Terraform Environment Variables  or create [terraform.tfvars file](https://www.terraform.io/language/values/variables#variable-definitions-tfvars-files)
    ```sh
    export TF_VAR_REGION=<aws region>
    ```
1. Apply Plan
    ```sh
    cd examples/npa/aws/existing-vpc    
    terraform init
    terraform apply
    ```
1. Destroy Plan
    ```sh
    terraform apply -destroy
    ```