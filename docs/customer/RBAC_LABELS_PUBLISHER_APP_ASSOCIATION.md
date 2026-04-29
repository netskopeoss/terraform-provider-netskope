

## Short Answer

Yes — the Terraform provider supports `label_ids` on both publishers and private apps. You can create an RBAC label like "BusinessUnitA", assign it to your publishers and private apps, and use that shared label in your Terraform configuration to control which publishers serve which apps.

## What Resources Support Labels

The provider supports `label_ids` on the following resources:

| Resource | Attribute |
|----------|-----------|
| `netskope_rbac_label` | Creates the label, outputs `label_id` |
| `netskope_npa_publisher` | `label_ids` — assigns labels to a publisher |
| `netskope_npa_private_app` | `label_ids` — assigns labels to a private app |
| `netskope_npa_local_broker` | `label_ids` — assigns labels to a local broker |

## How Labels Work

RBAC labels serve two purposes:

1. **Admin scoping in the Netskope UI** — an admin scoped to "BusinessUnitA" only sees publishers and apps tagged with that label. This enables delegated administration where different teams manage their own resources independently.

2. **Organizational key in Terraform** — you can use the shared label as a grouping mechanism in your Terraform code to automatically wire up publisher-to-app associations. The label gives you a single source of truth: "these publishers and these apps belong to BusinessUnitA."

It's important to note that labels don't automatically bind publishers to apps at the platform level. The `publishers` block on a private app is always an explicit assignment. What labels give you is a clean way to organize that assignment in your Terraform configuration, and scoped admin visibility in the UI as a bonus.

## Example: Your BusinessUnitA Scenario

This example shows exactly the pattern you described — a "BusinessUnitA" label shared between a publisher and private apps.

```hcl
# =============================================================================
# Step 1: Create the RBAC label
# =============================================================================

resource "netskope_rbac_label" "business_unit_a" {
  name  = "BusinessUnitA"
  color = "#0294C9"
}

# =============================================================================
# Step 2: Create publishers tagged with the label
# =============================================================================

resource "netskope_npa_publisher" "bu_a_east" {
  publisher_name = "BU-A-Publisher-East"
  label_ids      = [netskope_rbac_label.business_unit_a.label_id]
}

resource "netskope_npa_publisher" "bu_a_west" {
  publisher_name = "BU-A-Publisher-West"
  label_ids      = [netskope_rbac_label.business_unit_a.label_id]
}

# =============================================================================
# Step 3: Create private apps tagged with the same label
#         and assign the BusinessUnitA publishers
# =============================================================================

resource "netskope_npa_private_app" "bu_a_portal" {
  private_app_name     = "BU-A-Internal-Portal"
  private_app_hostname = "portal.businessunita.internal"
  label_ids            = [netskope_rbac_label.business_unit_a.label_id]

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.bu_a_east.publisher_id)
      publisher_name = netskope_npa_publisher.bu_a_east.publisher_name
    },
    {
      publisher_id   = tostring(netskope_npa_publisher.bu_a_west.publisher_id)
      publisher_name = netskope_npa_publisher.bu_a_west.publisher_name
    }
  ]

  use_publisher_dns = true
}

resource "netskope_npa_private_app" "bu_a_database" {
  private_app_name     = "BU-A-Database"
  private_app_hostname = "db.businessunita.internal"
  label_ids            = [netskope_rbac_label.business_unit_a.label_id]

  protocols = [
    {
      port     = "5432"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.bu_a_east.publisher_id)
      publisher_name = netskope_npa_publisher.bu_a_east.publisher_name
    },
    {
      publisher_id   = tostring(netskope_npa_publisher.bu_a_west.publisher_id)
      publisher_name = netskope_npa_publisher.bu_a_west.publisher_name
    }
  ]

  use_publisher_dns = true
}
```

After applying this configuration:

- Both publishers and both apps carry the "BusinessUnitA" label
- An admin scoped to "BusinessUnitA" sees only these resources in the Netskope console
- Both apps are served by both BU-A publishers

## Scaling to Multiple Business Units

If you have several business units, you can use `for_each` with a variable or local map to avoid repeating the label creation:

```hcl
variable "business_units" {
  default = {
    "BusinessUnitA" = "#0294C9"
    "BusinessUnitB" = "#FF5733"
    "BusinessUnitC" = "#28A745"
  }
}

resource "netskope_rbac_label" "bu" {
  for_each = var.business_units
  name     = each.key
  color    = each.value
}
```

Then reference the label when creating publishers and apps:

```hcl
resource "netskope_npa_publisher" "bu_b" {
  publisher_name = "BU-B-Publisher"
  label_ids      = [netskope_rbac_label.bu["BusinessUnitB"].label_id]
}

resource "netskope_npa_private_app" "bu_b_app" {
  private_app_name     = "BU-B-App"
  private_app_hostname = "app.businessunitb.internal"
  label_ids            = [netskope_rbac_label.bu["BusinessUnitB"].label_id]

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.bu_b.publisher_id)
      publisher_name = netskope_npa_publisher.bu_b.publisher_name
    }
  ]

  use_publisher_dns = true
}
```

## Multiple Labels on a Single Resource

A publisher or app can carry more than one label. For example, a publisher that serves both BusinessUnitA and a shared infrastructure label:

```hcl
resource "netskope_npa_publisher" "shared" {
  publisher_name = "Shared-Publisher"
  label_ids = [
    netskope_rbac_label.bu["BusinessUnitA"].label_id,
    netskope_rbac_label.bu["BusinessUnitB"].label_id,
  ]
}
```

Admins scoped to either label will see this publisher in the console.

## References

- [Label Based Access Control (RBAC V3)](https://docs.netskope.com/en/label-based-access-control-rbac-v3/) — Netskope documentation on RBAC labels and delegated administration
- [Private Access REST APIs](https://docs.netskope.com/en/private-access-rest-apis/) — API reference for publishers and private apps
