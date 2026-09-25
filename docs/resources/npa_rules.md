---
page_title: "netskope_npa_rules Resource - terraform-provider-netskope"
subcategory: "NPA"
description: |-
  Manages an NPA access rule that allows, blocks, or monitors user access to private applications.
---

# netskope_npa_rules (Resource)

Manages an NPA access rule. Rules are evaluated against user connections to determine whether access to a private application is allowed, blocked, or monitored.

## Rule Structure

Each rule has two logical parts inside `rule_data`:

- **Match criteria** — who the rule applies to: `users`, `user_groups`, `private_apps`, `device_classification_id`, `src_countries`, etc.
- **Match criteria action** — what happens when the criteria match: `allow`, `block`, or `periodic_reauth`. Configured via the `match_criteria_action` block.

## Policy Groups

Rules belong to a policy group (`group_id`). The **Default** group is shared with manually-created rules; use a dedicated [`netskope_npa_policy_groups`](npa_policy_groups) resource for Terraform-managed rules so they are isolated and their evaluation order is predictable.

## Rule Ordering

`rule_order` controls where a new rule is inserted in the list:

| `order` value | Behaviour |
|---|---|
| `"bottom"` | Added at the end (default) |
| `"top"` | Added at the beginning |
| `"before"` | Inserted before `rule_id` or `rule_name` |
| `"after"` | Inserted after `rule_id` or `rule_name` |

For large deployments managed with `for_each`, use [`netskope_npa_rules_order`](npa_rules_order) to set final positions in bulk after all rules are created.

## Targeting Private Apps

There are two ways to specify which private apps a rule applies to. They can be used together or independently:

- **`private_apps`** — a list of app names. The provider wraps each name in brackets automatically when communicating with the API; do not add brackets in config. Use the [`netskope_npa_private_app`](../data-sources/npa_private_app) data source to look up an app name by ID rather than hardcoding it:
  ```hcl
  data "netskope_npa_private_app" "example" {
    private_app_id = 123
  }

  resource "netskope_npa_rules" "example" {
    rule_data = {
      private_apps = [data.netskope_npa_private_app.example.private_app_name]
      ...
    }
  }
  ```
- **`private_app_tag_ids`** — a list of tag IDs (as strings, e.g. `["1542"]`). The rule applies to all apps carrying any of those tags. Tag IDs are numeric and can be found in the Netskope UI under Settings → Private Apps → Tags. There is no API endpoint to list them programmatically.

`private_app_tags` (the tag names corresponding to `private_app_tag_ids`) is **read-only** and populated automatically by the API. Do not set it in config — set `private_app_tag_ids` instead.

## Device Classification Criteria

Two distinct fields control device posture matching, and they work differently:

- **`classification`** — matches against the built-in Netskope device posture categories. Valid values are `"managed"` and `"unmanaged"`. Use this for broad rules that apply to all managed or all unmanaged devices regardless of specific tag assignments.
- **`device_classification_id`** — matches against specific custom device classification tags created in the Netskope UI. Takes a list of tag IDs as strings. Use `tostring(tag.tag_id)` when referencing a [`netskope_device_classification_tag`](device_classification_tag) data source.

These fields can be used independently or together depending on the required match logic.

## Block Rules

To block access and show a block page, set `action_name = "block"` and provide a `template` display name (e.g. `"Default Template"`). The `template` value **must be the display name** shown in the Netskope UI under Settings → NPA → Block Templates — not the underlying file name (e.g. `"1.html"`).

## Schedule

`schedule` restricts when a rule is active. Each schedule entry supports two mutually exclusive methods for defining the active window — use one or the other within each entry, not both:

- **`time_interval_obj`** — a list of Time Interval object IDs configured in the Netskope console (Policies → Time Intervals). IDs are numeric strings and must be looked up in the UI; there is no API endpoint to list them.
- **`time_range`** — an explicit list of date/time windows, each with `start_date`, `start_time`, `end_date`, and `end_time` in `MM/DD/YYYY` and `HH:MM` format.

When using only `time_interval_obj`, set `time_range = []` explicitly in the same schedule entry. If `time_range` is omitted, the provider may show it as unknown on subsequent plans.

Once a schedule is set, it cannot be cleared by removing `schedule` from config — the provider preserves the existing schedule in state to avoid phantom diffs. To change the schedule, update it to the new value rather than removing it.

## Periodic Re-authentication

`periodic_reauth` forces users to re-authenticate after a set interval. It is only meaningful on `allow` rules. Set `reauth_interval` as a quoted number (e.g. `"60"`) and `reauth_interval_unit` as `"hours"` or `"days"`.

## Known Limitations

**Array fields cannot be cleared by omitting them from config.** Fields such as `users`, `user_groups`, `src_countries`, and `schedule` are preserved in state when absent from config to avoid phantom plan diffs. This means removing one of these fields from your HCL will not send an empty value to the API — the existing values are retained. To effectively remove criteria, update the field to a new value rather than omitting it entirely.

## Common Mistakes

- **`enabled = true`** causes a type error — the field must be the string `"1"` (enabled) or `"0"` (disabled).
- **Private app names must not be wrapped in brackets** — use `"my-app"` not `"[my-app]"`. The provider adds brackets automatically when communicating with the API.
- **`device_classification_id`** takes numeric tag IDs as strings. Use `tostring(tag.tag_id)` when referencing a `netskope_device_classification_tag` data source.
- **`private_app_tags` is read-only** — do not set it in config. Set `private_app_tag_ids` instead; the provider populates `private_app_tags` automatically from the API response.
- **`time_range = []` must be set explicitly** in schedule entries that only use `time_interval_obj`, otherwise the provider may show `time_range` as unknown on subsequent plans.

## Example Usage

```terraform
resource "netskope_npa_rules" "my_nparules" {
  description = "any"
  enabled     = "1"
  group_id    = "1"
  rule_data = {
    access_method = [
      "Clientless"
    ]
    b_negate_net_location  = false
    b_negate_src_countries = false
    classification = [
      "..."
    ]
    description = "...my_description..."
    device_classification_id = [
      "..."
    ]
    json_version = 3
    match_criteria_action = {
      action_name = "allow"
      emit_alert  = true
      template    = "...my_template..."
    }
    net_location_obj = [
      "27",
      "42",
    ]
    notify = {
      emails = [
        "..."
      ]
      from_user = "...my_from_user..."
      interval  = "30"
      to_users = [
        "..."
      ]
    }
    organization_units = [
      "engineering/qa",
    ]
    os = [
      "..."
    ]
    periodic_reauth = {
      reauth_interval      = "60"
      reauth_interval_unit = "hours"
    }
    policy_type = "private-app"
    private_app_tag_ids = [
      "1",
      "2",
    ]
    private_app_tags = [
      "tag1",
      "tag2",
    ]
    private_apps = [
      "app1",
      "app2",
    ]
    schedule = [
      {
        time_interval_obj = [
          "..."
        ]
        time_range = [
          {
            end_date   = "MM/DD/YYYY"
            end_time   = "HH:MM"
            start_date = "MM/DD/YYYY"
            start_time = "HH:MM"
          }
        ]
      }
    ]
    src_countries = [
      "US",
      "AF",
      "CN",
    ]
    user_confidence = {
      index    = "351"
      operator = "lt"
    }
    user_groups = [
      "usergroup/group1",
    ]
    user_type = "user"
    users = [
      "vphan@netskope.com",
    ]
    version = 1
  }
  rule_name = "vantest"
  rule_order = {
    order     = "before"
    position  = 5
    rule_id   = 1
    rule_name = "api-policy-managed"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `description` (String)
- `enabled` (String)
- `group_id` (String)
- `rule_data` (Attributes) (see [below for nested schema](#nestedatt--rule_data))
- `rule_name` (String)
- `rule_order` (Attributes) (see [below for nested schema](#nestedatt--rule_order))

### Read-Only

- `group_name` (String) Policy group name this rule belongs to (read-only, returned by API)
- `id` (String) policy rule id

<a id="nestedatt--rule_data"></a>
### Nested Schema for `rule_data`

Optional:

- `access_method` (List of String) Default: []
- `b_negate_net_location` (Boolean) Default: false
- `b_negate_src_countries` (Boolean) Default: false
- `classification` (List of String) Device classification filter: list of managed/unmanaged categories to match (e.g. ["unmanaged"]). Set in the Netskope UI under Device Classification criteria. Default: []
- `description` (String) Description stored within rule_data (separate from the top-level rule description)
- `device_classification_id` (List of String) Default: []
- `json_version` (Number) Default: 3
- `match_criteria_action` (Attributes) (see [below for nested schema](#nestedatt--rule_data--match_criteria_action))
- `net_location_obj` (List of String) List of Network Location IDs to match. Network Locations are defined in the Netskope tenant UI (Policies > Network Locations) and referenced here by their numeric ID (e.g. "27"). Default: []
- `notify` (Attributes) Notification configuration for alert/block rule actions (see [below for nested schema](#nestedatt--rule_data--notify))
- `organization_units` (List of String) Default: []
- `os` (List of String) Operating system filter (Client access only). Valid values: "AmigaOS", "Android", "BlackBerry", "BSD", "Chrome OS", "Darwin", "Debian", "Fedora", "iOS", "Linux", "Mac", "Others", "Red", "RHEL", "Solaris", "SunOS", "Symbian", "Ubuntu", "Windows". Default: []
- `periodic_reauth` (Attributes) (see [below for nested schema](#nestedatt--rule_data--periodic_reauth))
- `policy_type` (String) Default: "private-app"; must be "private-app"
- `private_app_tag_ids` (List of String) Tag IDs (numeric as string) — alternative to privateAppTags (names). Default: []
- `private_app_tags` (List of String) Default: []
- `private_apps` (List of String) Default: []
- `schedule` (Attributes List) Schedule configuration for policy enforcement timing (see [below for nested schema](#nestedatt--rule_data--schedule))
- `src_countries` (List of String) Default: []
- `user_confidence` (Attributes) User Confidence Index filter. Requires the User Confidence Index feature to be enabled on the tenant. (see [below for nested schema](#nestedatt--rule_data--user_confidence))
- `user_groups` (List of String) Default: []
- `user_type` (String) Default: "user"; must be "user"
- `users` (List of String) Default: []
- `version` (Number)

<a id="nestedatt--rule_data--match_criteria_action"></a>
### Nested Schema for `rule_data.match_criteria_action`

Optional:

- `action_name` (String) must be one of ["allow", "block"]
- `emit_alert` (Boolean) Whether to emit an alert when the rule matches (required for block action)
- `template` (String) Notification template name (required for block action). Use the display name (e.g. "Default Template"), not the file name.


<a id="nestedatt--rule_data--notify"></a>
### Nested Schema for `rule_data.notify`

Optional:

- `emails` (List of String) Email addresses to notify
- `from_user` (String) Sender user identifier
- `interval` (String) Notification interval in minutes (as string, e.g. '30')
- `to_users` (List of String) Recipient user types (e.g. 'admin')


<a id="nestedatt--rule_data--periodic_reauth"></a>
### Nested Schema for `rule_data.periodic_reauth`

Optional:

- `reauth_interval` (String)
- `reauth_interval_unit` (String)


<a id="nestedatt--rule_data--schedule"></a>
### Nested Schema for `rule_data.schedule`

Optional:

- `time_interval_obj` (List of String) IDs of Time Interval objects configured in the Netskope console (Policies > Time Intervals). No API endpoint exists to list these; obtain IDs from the Netskope UI. Default: []
- `time_range` (Attributes List) Date/time ranges when the policy is active (see [below for nested schema](#nestedatt--rule_data--schedule--time_range))

<a id="nestedatt--rule_data--schedule--time_range"></a>
### Nested Schema for `rule_data.schedule.time_range`

Optional:

- `end_date` (String)
- `end_time` (String)
- `start_date` (String)
- `start_time` (String)



<a id="nestedatt--rule_data--user_confidence"></a>
### Nested Schema for `rule_data.user_confidence`

Optional:

- `index` (String) Confidence index threshold value (e.g. 350, 351, 650, 651)
- `operator` (String) Comparison operator: lt (below threshold) or gt (above threshold)



<a id="nestedatt--rule_order"></a>
### Nested Schema for `rule_order`

Optional:

- `order` (String) must be one of ["top", "bottom", "before", "after"]
- `position` (Number)
- `rule_id` (Number)
- `rule_name` (String)

## Import

NPA rules are imported by their numeric rule ID.

| Approach | Terraform version | Config generation |
|----------|------------------|-------------------|
| **`import` block + `-generate-config-out`** | ≥ 1.5 | Automatic — Terraform writes the HCL |
| **`terraform import` CLI** | Any | Manual — copy values from `terraform state show` |

Use the `import` block approach whenever possible. It generates complete HCL from the API response so you do not need to copy field values by hand.

### Finding rule IDs

Every NPA rule has a numeric ID assigned by the API:

```bash
curl "https://<tenant>.goskope.com/api/v2/policy/npa/rules" \
  -H "Netskope-Api-Token: <token>" | jq '.data[] | {id: .rule_id, name: .rule_name}'
```

Or look up IDs in Terraform using the data source:

```hcl
data "netskope_npa_rules_list" "all" {}

output "rule_ids" {
  value = { for r in data.netskope_npa_rules_list.all.rules : r.rule_name => r.id }
}
```

### Approach 1: `import` block with automatic config generation (Terraform ≥ 1.5)

**Step 1** — Add one `import` block per rule. Do not write `resource` blocks yet — they are generated in the next step.

```hcl
import {
  to = netskope_npa_rules.web_allow
  id = "<id>"
}

import {
  to = netskope_npa_rules.db_block
  id = "<id>"
}
```

**Step 2** — Generate resource config automatically:

```bash
terraform plan -generate-config-out=generated.tf
```

Terraform reads each rule from the API and writes a complete `resource` block into `generated.tf` with every field filled in:

```hcl
# __generated__ by Terraform from "<id>"
resource "netskope_npa_rules" "web_allow" {
  description = null
  enabled     = "1"
  group_id    = null
  rule_data = {
    access_method            = ["Client"]
    b_negate_net_location    = false
    b_negate_src_countries   = false
    classification           = []
    description              = null
    device_classification_id = ["22688"]
    json_version             = 3
    match_criteria_action = {
      action_name = "allow"
      emit_alert  = null
      template    = null
    }
    net_location_obj    = []
    organization_units  = []
    os                  = ["Mac"]
    periodic_reauth     = null
    policy_type         = "private-app"
    private_app_tag_ids = []
    private_app_tags    = []
    private_apps        = ["my-internal-app"]
    schedule = [
      {
        time_interval_obj = ["14"]
        time_range        = []
      },
    ]
    src_countries   = []
    user_confidence = null
    user_groups     = ["engineering-team"]
    user_type       = "user"
    users           = ["alice@example.com"]
  }
  rule_name  = "web-allow-client"
  rule_order = null
}
```

**Step 3** — Review `generated.tf`. It is valid as-is. Optionally clean up `null` values for fields you don't need to manage — they are preserved from state automatically when absent from config:

| Field | Safe to remove |
|-------|---------------|
| `description = null` | Preserved from state when absent |
| `group_id = null` | API doesn't return it on GET; no drift |
| `rule_order = null` | Only meaningful on create |
| `emit_alert = null` | Only needed when emitting alerts |
| `template = null` | Only needed for block / periodic_reauth actions |
| `periodic_reauth = null` | Only needed when action is periodic_reauth |
| `user_confidence = null` | Only needed when using UCI feature |
| `b_negate_net_location = false` | Default value |
| `b_negate_src_countries = false` | Default value |
| `json_version = 3` | Computed by API |
| `policy_type = "private-app"` | Default value |
| `user_type = "user"` | Default value |
| Empty lists (`= []`) | Default value |

**Keep all non-empty list values.** Lists like `private_apps`, `access_method`, `device_classification_id`, `user_groups`, `users`, and `schedule` must remain in config if non-empty — removing them causes Terraform to plan clearing them.

**Step 4** — Move `generated.tf` content into your main config, then apply:

```bash
terraform apply
terraform plan   # expected: No changes
```

### Approach 2: `terraform import` CLI (any Terraform version)

**Step 1** — Write placeholder resource blocks. Use `rule_data = {}` — a completely empty block (no `rule_data`) causes all fields to show as `(known after apply)`:

```hcl
resource "netskope_npa_rules" "web_allow" {
  rule_data = {}
}
```

**Step 2** — Import by numeric ID:

```bash
terraform import netskope_npa_rules.web_allow <id>
```

**Step 3** — Inspect state to get all field values:

```bash
terraform state show netskope_npa_rules.web_allow
```

Output example:

```
resource "netskope_npa_rules" "web_allow" {
    enabled   = "1"
    id        = "<id>"
    rule_data = {
        access_method            = ["Client"]
        device_classification_id = ["22688"]
        match_criteria_action    = {
            action_name = "allow"
        }
        os                       = ["Mac"]
        private_apps             = ["my-internal-app"]
        schedule                 = [{
            time_interval_obj = ["14"]
            time_range        = []
        }]
        user_groups              = ["engineering-team"]
        users                    = ["alice@example.com"]
    }
    rule_name = "web-allow-client"
}
```

Every value shown — group names, device classification IDs, time interval IDs — was read from the API during import. Copy them verbatim into your config.

**Step 4** — Replace the placeholder block with a full config from the state output, include all non-empty lists, then verify:

```bash
terraform plan   # expected: No changes
```

### Import troubleshooting

| Plan shows | Cause | Fix |
|-----------|-------|-----|
| Removing items from a list | Non-empty list in state but absent from config | Add the list and its values to config |
| `template` drift | Display name / `.html` filename mismatch | Use the display name exactly as shown in state |
| `(known after apply)` on all fields | `rule_data` omitted from config block | Add `rule_data = {}` as minimum placeholder |
| `schedule` drift | `time_range` omitted | Include `time_range = []` when using `time_interval_obj` |

### Import examples

#### Simple allow rule

```hcl
resource "netskope_npa_rules" "web_allow" {
  rule_name = "web-allow-client"
  enabled   = "1"

  rule_data = {
    match_criteria_action = {
      action_name = "allow"
    }
    private_apps  = ["my-web-app"]
    access_method = ["Client"]
  }
}
```

#### Block rule with notification template

The `template` value is the display name as shown in the Netskope UI — the provider translates `.html` filenames to display names automatically on import and refresh.

```hcl
resource "netskope_npa_rules" "db_block" {
  rule_name = "db-block-unmanaged"
  enabled   = "1"

  rule_data = {
    match_criteria_action = {
      action_name = "block"
      template    = "Default Block Page"
    }
    private_apps             = ["my-db-app"]
    access_method            = ["Client"]
    os                       = ["Windows"]
    device_classification_id = ["22691"]
    user_groups              = ["all-employees"]
    users                    = ["contractor@example.com"]
    schedule = [{
      time_interval_obj = ["14"]
      time_range        = []
    }]
  }
}
```

#### Periodic re-authentication rule

```hcl
resource "netskope_npa_rules" "app_reauth_daily" {
  rule_name = "app-reauth-daily"
  enabled   = "1"

  rule_data = {
    match_criteria_action = {
      action_name = "periodic_reauth"
      template    = "Re-authentication Required"
    }
    private_apps             = ["my-internal-app"]
    access_method            = ["Client"]
    os                       = ["Mac"]
    device_classification_id = ["22688"]
    user_groups              = ["engineering-team"]
    users                    = ["alice@example.com"]
    periodic_reauth = {
      reauth_interval      = "1"
      reauth_interval_unit = "days"
    }
    schedule = [{
      time_interval_obj = ["14"]
      time_range        = []
    }]
  }
}
```

#### Rule with OU and network location filter

A Clientless rule scoped to an Active Directory OU, a specific network location, and source country.

```hcl
resource "netskope_npa_rules" "ou_web_allow" {
  rule_name = "ou-web-allow-clientless"
  enabled   = "1"

  rule_data = {
    match_criteria_action = {
      action_name = "allow"
    }
    private_apps       = ["my-web-app"]
    access_method      = ["Clientless"]
    organization_units = ["Corp/Engineering"]
    net_location_obj   = ["4749e572-604d-45dc-bac5-2d3fc3cce732"]
    src_countries      = ["US"]
    schedule = [{
      time_interval_obj = ["14"]
      time_range        = []
    }]
  }
}
```

The `net_location_obj` UUID comes from your tenant's Network Locations configuration. After import it appears in state — copy it verbatim. To look it up:

```bash
curl "https://<tenant>.goskope.com/api/v2/policy/network-locations" \
  -H "Netskope-Api-Token: <token>" | jq '.data[] | {id: .id, name: .name}'
```

#### Source country filter

```hcl
resource "netskope_npa_rules" "geo_allow" {
  rule_name = "geo-allow-emea"
  enabled   = "1"

  rule_data = {
    match_criteria_action = {
      action_name = "allow"
    }
    private_apps  = ["my-web-app"]
    access_method = ["Client"]
    src_countries = ["GB", "DE", "FR", "NL"]
  }
}
```

#### Importing multiple rules at once

```hcl
# imports.tf
import { to = netskope_npa_rules.web_allow    id = "<id>" }
import { to = netskope_npa_rules.db_block     id = "<id>" }
import { to = netskope_npa_rules.app_reauth   id = "<id>" }
import { to = netskope_npa_rules.ou_web_allow id = "<id>" }
```

```bash
terraform plan -generate-config-out=generated.tf
terraform apply
terraform plan   # expected: No changes
```

### Notes on specific fields

**`notify` is not imported** — email notification configuration is computed by the API from the rule's action and template settings. It is excluded from the Terraform schema and always shows as `null` in state. Configure it via the notification template, not Terraform.

**`group_id` is not returned by the API on read** — the API does not include the policy group ID in GET responses. After import, `group_id` is `null` in state. This does not cause drift. Set it in config if you want Terraform to enforce group membership.

**`userGroupObjects` is computed and ignored** — the API enriches `user_groups` with full group detail in a `userGroupObjects` field. This is excluded from the Terraform schema and does not appear in state.

**Template names** — on import and every refresh, the provider translates `.html` template filenames returned by the API to display names. State always stores the display name. Use the value shown by `terraform state show` or in `generated.tf` verbatim — never the `.html` filename.