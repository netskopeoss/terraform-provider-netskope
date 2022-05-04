# Netskope Publisher and Private App Creation
This is an example plan that uses the Netskope Provider to create  2 Publishers and a Private App inside a Netskope Tenant.

[Module Documentation](./Module.md)

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
1. Run Simple Plan
    ```sh
    cd examples/npa/simple
    terraform init
    terraform apply
    ```
1. Destroy Simple Objects
    ```sh
    terraform apply -destroy
    ```