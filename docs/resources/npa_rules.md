---
page_title: "netskope_npa_rules Resource - terraform-provider-netskope"
subcategory: ""
description: |-
  Manages an NPA access rule that allows, blocks, or monitors user access to private applications.
---

# netskope_npa_rules (Resource)

Manages an NPA access rule. Rules are evaluated against user connections to determine whether access to a private application is allowed, blocked, or monitored.

## Rule Structure

Each rule has two logical parts inside `rule_data`:

- **Match criteria** — who the rule applies to: `users`, `user_groups`, `private_apps`, `device_classification_id`, `src_countries`, etc.
- **`match_criteria_action`** — what happens when the criteria match: `allow`, `block`, or `monitor`.

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

## Notifications

`notify` sends an alert when the rule matches. Typically used on `block` or `monitor` rules. `interval` is the notification frequency in minutes, supplied as a quoted number (e.g. `"60"`). `to_users` accepts role types such as `"admin"`.

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
      "user@example.com",
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

Import is supported using the following syntax:

```shell
terraform import netskope_npa_rules.my_netskope_npa_rules "1"
```