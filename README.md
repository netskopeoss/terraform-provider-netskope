# Terraform Provider Netskope




## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 1.1.7
-	[Go](https://golang.org/doc/install) >= 1.17
-   [netskope-api-client-go](https://github.com/netskopeoss/netskope-api-client-go) 


## Building The Provider - on a Mac

1. Clone netskope-api-client-go Git Repo
1. Clone Netskope Provider Repo
1. Navigate to the Netskope Provider Dir
1. Compile the Netskope Provider

```sh
git clone https://github.com/netskopeoss/netskope-api-client-go.git
git clone https://github.com/netskopeoss/terraform-provider-netskope.git
cd terraform-provider-netskope
make install
```


## Using  The Provider
The Netskope Terraform Provider Repo includes sample plans to get you started. You will need to complete several task before launching any of the samples or to use the provider in your own plans.

### Netskope Tenant Tasks

1. Identify the "Base URL" for your Netskope tenant.
    - This will be the URL used to manage your Netskope tenant 
    - For example: `https://example.goskope.com`
1. Follow the [REST APIv2 Documentaion](https://docs.netskope.com/en/rest-api-v2-overview-312207.html) to create an API Token
    - Required "Read+Write" Endpoints For NPA:
        - `/api/v2/steering/private`
        - `/api/v2/infrastructure/publisher`
    ![API Token](images/npa_api_token.png)

### Simple Terraform Plan
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
    cd examples/simple
    terraform init
    terraform apply
    ```
1. Destroy Simple Objects
    ```sh
    terraform apply -destroy
    ```



`