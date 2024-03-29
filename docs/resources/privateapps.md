---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netskope_privateapps Resource - terraform-provider-netskope"
subcategory: ""
description: |-
  
---

# netskope_privateapps (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_name` (String)
- `host` (String)
- `protocols` (Block List, Min: 1) (see [below for nested schema](#nestedblock--protocols))

### Optional

- `clientless_access` (Boolean)
- `publisher` (Block List) (see [below for nested schema](#nestedblock--publisher))
- `tags` (Block List) (see [below for nested schema](#nestedblock--tags))
- `trust_self_signed_certs` (Boolean)
- `use_publisher_dns` (Boolean)

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--protocols"></a>
### Nested Schema for `protocols`

Required:

- `port` (String)
- `type` (String)


<a id="nestedblock--publisher"></a>
### Nested Schema for `publisher`

Required:

- `publisher_id` (String)
- `publisher_name` (String)


<a id="nestedblock--tags"></a>
### Nested Schema for `tags`

Required:

- `tag_name` (String)
