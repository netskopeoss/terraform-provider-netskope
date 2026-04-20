# Device Classification in NPA Rules

Device classification tags let you enforce policies based on device posture. For example, you can allow access only from devices that have CrowdStrike installed, or restrict unmanaged devices.

Tags are created in the Netskope UI under **Settings > Device Classification**. Each tag has a numeric ID. The `netskope_device_classification_tag_list` data source lets you look up these IDs from Terraform instead of hardcoding them.

## Data Source Schema

| Attribute      | Type         | Description                                      |
|----------------|--------------|--------------------------------------------------|
| `tag_id`       | number       | Numeric ID (used in `device_classification_id`)  |
| `name`         | string       | Tag name as shown in the UI                      |
| `description`  | string       | Tag description                                  |
| `priority`     | number       | Priority (lower values = higher priority)        |
| `policy_names` | list(string) | Policies currently using this tag                |

## Example 1: List All Tags

Retrieve all device classification tags and output them for reference.

```hcl
data "netskope_device_classification_tag_list" "all" {}

output "device_classification_tags" {
  description = "All device classification tags with their IDs"
  value = [
    for t in data.netskope_device_classification_tag_list.all.tags : {
      id          = t.tag_id
      name        = t.name
      description = t.description
      priority    = t.priority
    }
  ]
}
```

After `terraform apply`, the output shows all available tags:

```
device_classification_tags = [
  { id = 8505, name = "CrowdStrike ZTA Low Risk", description = "", priority = 5 },
  { id = 9117, name = "CrowdStrike Installed",    description = "", priority = 3 },
  { id = 9120, name = "SentinelOne Installed",     description = "", priority = 2 },
]
```

## Example 2: Look Up a Tag by Name

Use a `for` expression to find a specific tag's ID by name.

```hcl
data "netskope_device_classification_tag_list" "all" {}

locals {
  crowdstrike_tag_id = tostring([
    for t in data.netskope_device_classification_tag_list.all.tags : t.tag_id
    if t.name == "CrowdStrike Installed"
  ][0])
}
```

## Example 3: Restrict an NPA Rule to Classified Devices

Create a rule that only allows access from devices matching a specific classification.

```hcl
data "netskope_device_classification_tag_list" "all" {}

locals {
  crowdstrike_tag_id = tostring([
    for t in data.netskope_device_classification_tag_list.all.tags : t.tag_id
    if t.name == "CrowdStrike Installed"
  ][0])
}

resource "netskope_npa_rules" "crowdstrike_only" {
  rule_name = "allow-crowdstrike-devices"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]

    device_classification_id = [local.crowdstrike_tag_id]
  }

  rule_order = {
    order = "top"
  }
}
```

## Example 4: Multiple Classifications on a Single Rule

You can pass multiple tag IDs to `device_classification_id`. The device must match **any** of the listed classifications.

```hcl
data "netskope_device_classification_tag_list" "all" {}

locals {
  tag_ids_by_name = { for t in data.netskope_device_classification_tag_list.all.tags : t.name => t.tag_id }
}

resource "netskope_npa_rules" "edr_required" {
  rule_name = "allow-edr-devices"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]

    device_classification_id = [
      tostring(local.tag_ids_by_name["CrowdStrike Installed"]),
      tostring(local.tag_ids_by_name["SentinelOne Installed"]),
    ]
  }
}
```

## Example 5: Validate a Tag Exists

Use a `precondition` to fail early if an expected tag is missing from the tenant.

```hcl
data "netskope_device_classification_tag_list" "all" {}

locals {
  required_tag  = "CrowdStrike Installed"
  matching_tags = [for t in data.netskope_device_classification_tag_list.all.tags : t if t.name == local.required_tag]
}

resource "netskope_npa_rules" "validated" {
  rule_name = "allow-validated-devices"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  lifecycle {
    precondition {
      condition     = length(local.matching_tags) > 0
      error_message = "Device classification tag '${local.required_tag}' not found. Create it in the Netskope UI under Settings > Device Classification."
    }
  }

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]

    device_classification_id = [tostring(local.matching_tags[0].tag_id)]
  }
}
```

## Notes

- **Tag IDs are integers** but `device_classification_id` expects strings in Terraform config. Use `tostring()` to convert.
- **The provider automatically converts** string IDs to integers when sending to the API.
- **Tags are created in the Netskope UI**, not via Terraform. This data source is read-only.
- **"Managed" and "Unmanaged"** are not built-in tags. They are custom device classification labels that administrators create. Once created, they appear in this data source like any other tag.

## Test Results

The data source is backed by acceptance tests that verify real tag data is returned from the API. Tests passed on 2026-04-20:

```
=== RUN   TestAccDeviceClassificationTagListDataSource_basic
--- PASS: TestAccDeviceClassificationTagListDataSource_basic (3.54s)
PASS
```
