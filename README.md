# Terraform Provider Netskope




## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 1.1.7
-	[Go](https://golang.org/doc/install) >= 1.17
-   [NSGO](https://github.com/ns-sbrown/nsgo) 


## Building The Provider - on a Mac

1. Clone NSGO Git Repo
1. Clone Netskope Provider Repo
1. Navigate to the Netskope Provider Dir
1. Compile the Netskope Provider

```sh
git clone https://github.com/ns-sbrown/nsgo.git
git clone https://github.com/ns-sbrown/terraform-provider-netskope.git
cd terraform-provider-netskope
make install
```


## Using  The Provider
1. Export Provider Env Vsriables ```sh export NS_BaseURL=https://sedemo.goskope.com && export NS_ApiToken=1234567890```
1. Something Else



`