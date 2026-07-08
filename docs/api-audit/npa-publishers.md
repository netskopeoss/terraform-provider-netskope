# API Audit: NPA Publishers

**Date:** 2026-07-07
**Tenant:** bespin.goskope.com
**Base path:** `/api/v2/infrastructure/publishers`

## Endpoints Tested

| Method | Path | Status | Notes |
|--------|------|--------|-------|
| GET | `/infrastructure/publishers` | 200 | Returns all publishers |
| GET | `/infrastructure/publishers/{id}` | 200 | Returns single publisher by ID |

---

## `connected_apps` — Shape Differs Between List and Get-by-ID

### What the Swagger says

The OAS defines `connected_apps` as `array of strings` in both the list schema
(`publishers_get_response`) and the get-by-ID schema (`publisher_response`):

```yaml
connected_apps:
  type: array
  items:
    type: string
  example:
    - '[Cloud Exchange]'
    - '[WebServer]'
```

### What the API actually returns

**`GET /infrastructure/publishers`** — matches the Swagger: array of strings.

```json
{
    "data": {
        "publishers": [
            {
                "apps_count": 1,
                "assessment": null,
                "capabilities": null,
                "common_name": "3418b5ec6ad1f703",
                "connected_apps": [
                    "[BD-Gartner-HPE]"
                ],
                "last_download_url": null,
                "last_log_collection_status": null,
                "last_log_collection_timestamp": null,
                "last_log_s3_key": null,
                "lbrokerconnect": false,
                "log_collection_in_progress": false,
                "publisher_id": 10950,
                "publisher_name": "BD-gartner-hpe-sd-wan",
                "publisher_upgrade_profiles_external_id": 1,
                "registered": false,
                "status": "not registered",
                "stitcher_id": null,
                "stitcher_pop": null,
                "tags": [],
                "upgrade_failed_reason": null,
                "upgrade_request": false,
                "upgrade_status": {
                    "upstat": "not_support"
                }
            }
        ]
    },
    "status": "success",
    "total": 11
}
```

**`GET /infrastructure/publishers/10950`** — does NOT match the Swagger: returns an
array of objects, not strings.

```json
{
    "data": {
        "apps_count": 1,
        "assessment": null,
        "capabilities": null,
        "common_name": "3418b5ec6ad1f703",
        "connected_apps": [
            {
                "access_method": "client",
                "host": "172.20.36.40",
                "last_connected": null,
                "name": "[BD-Gartner-HPE]"
            }
        ],
        "id": 10950,
        "labels": [],
        "last_download_url": null,
        "last_log_collection_status": null,
        "last_log_collection_timestamp": null,
        "last_log_s3_key": null,
        "lbrokerconnect": false,
        "log_collection_in_progress": false,
        "name": "BD-gartner-hpe-sd-wan",
        "publisher_upgrade_profiles_id": 1,
        "registered": false,
        "status": "not registered",
        "stitcher_id": null,
        "stitcher_pop": null,
        "tags": [],
        "upgrade_failed_reason": null,
        "upgrade_request": false,
        "upgrade_status": {
            "upstat": "not_support"
        }
    },
    "status": "success"
}
```

### Summary of deviations

| Field | Swagger (publisher_response) | GET /publishers (list) | GET /publishers/{id} |
|-------|------------------------------|------------------------|----------------------|
| `connected_apps` items | `string` | `string` ✓ | `object` ✗ |
| `connected_apps[].access_method` | not defined | n/a | `string` |
| `connected_apps[].host` | not defined | n/a | `string` |
| `connected_apps[].last_connected` | not defined | n/a | `string\|null` |
| `connected_apps[].name` | not defined | n/a | `string` |

### Additional field name differences between list and get-by-ID

The two endpoints also use different field names for the same values:

| Concept | GET /publishers (list) | GET /publishers/{id} |
|---------|------------------------|----------------------|
| Publisher ID | `publisher_id` | `id` |
| Publisher name | `publisher_name` | `name` |
| Upgrade profile ID | `publisher_upgrade_profiles_external_id` | `publisher_upgrade_profiles_id` |
| Labels | not present | `labels: []` |

### Provider impact

The Swagger deviation in `connected_apps` : the Go SDK struct
declared `ConnectedApps []string`, which matched the list endpoint but panicked
on the get-by-ID response for any publisher with connected apps:

```
json: cannot unmarshal object into Go value of type string
```

**Fix:** `connected_apps` is annotated `x-speakeasy-ignore: true` in
`publisher_response` in the OAS. The field is dropped from the SDK struct
entirely and the JSON decoder skips it. The field was already excluded from
Terraform state (`x-speakeasy-terraform-ignore`), so there is no change in
provider behaviour.