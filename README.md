# Terraform Provider Netskope




## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 1.1.7
-	[Go](https://golang.org/doc/install) >= 1.17
-   [NSGO](https://github.com/ns-sbrown/nsgo) 


## Building The Provider - on a Mac

1. Clone NSGO Git Repo
```sh
git clone https://github.com/ns-sbrown/nsgo.git
```
1. Clone Netskope Provider Repo
```sh
git clone https://github.com/ns-sbrown/terraform-provider-netskope.git
```
1. Navigate to the Netskope Provider Dir
```sh
cd terraform-provider-netskope
```
1. Compile the Netskope Provider
```
make install
```
`