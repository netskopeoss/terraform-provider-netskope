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
1. Export Provider Env Variables 
    ```sh 
    export NS_BaseURL=https://sedemo.goskope.com && export NS_ApiToken=1234567890
    ```
1. Run Simple Plan
    ```sh
    cd examples/simple
    terraform init && terraform apply
    ```
1. Destroy Simple Objects
    ```sh
    terraform apply -destroy
    ```



`