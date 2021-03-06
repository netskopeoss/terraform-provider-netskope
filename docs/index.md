---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "Netskope Provider"
subcategory: ""
description: |-
Terraform Provider for Netskope APIv2 Endpoints. Currently supports Netskope Private Access Endpoints including Publishers and Private Apps.
---

# Netskope Provider

The Netskope Provider is used to manage resources in the Netskope Tenant. You can create, read, update and delete supported resources, which include Private Access Publishers and Applications. 


## Netskope Authentication

There are 2 options available for providing credentials to Terraform. These are to use Environment Variables or Provider block. It is strongly encourage the use of environment variables as the Provider block risks exposing the credentials publicly.

- Environment Variables(Recommended)
    - Mac or Linux
    ```sh
        export NS_BaseURL=<Base URL>
        export NS_ApiToken<API Token>`
    ```
    - Windows
    ```cmd
        set NS_BaseURL=<Base URL>
        set NS_ApiToken<API Token>
    ```

- Provider Block - Recommend using Environment variables instead.
    ```go
        provider "netskope" {
            baseurl = "https://<tenant-url>.goskope.com"
            apitoken = "<api token>"
        }
    ```

## Example Usage

```go
terraform {
  required_providers {
    netskope = {
      version = "0.1.0"
      source  = "github.com/netskopeoss/netskope"
    }
  }
}

resource "netskope_publishers" "Publisher" {
  name = "ExamplePublisher"
}


resource "netskope_privateapps" "PrivateApp" {
  app_name = "ExampleApp"
  host     = "site1.example-app.com, site2.example-app.com"

  protocols {
    type = "tcp"
    port = "22, 443, 8080-8081" //Ports can be single port or ranges
  }

  protocols {
    type = "udp" 
    port = "123"
  }

  publisher {
    publisher_id   = netskope_publishers.Publisher.id
    publisher_name = netskope_publishers.Publisher.name
  }
}
```



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `apitoken` (String, Sensitive) [Netskope REST APIv2 Token](https://docs.netskope.com/en/rest-api-v2-overview-312207.html)
- `baseurl` (String) Netskope Tenant URL i.e. example.goskope.com
