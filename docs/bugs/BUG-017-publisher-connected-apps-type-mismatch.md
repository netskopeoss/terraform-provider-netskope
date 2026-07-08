---
title: BUG-017 — Publisher ReadResource fails due to connected_apps type mismatch
status: Fixed
github_issue: 96
---

## Summary

`netskope_npa_publisher` ReadResource fails with:

```
error unmarshaling json response body: json: cannot unmarshal object into Go value of type string
```

This occurs during `terraform plan` or `terraform apply` for **existing publishers that have connected apps**.

## Root Cause

The `GET /api/v2/infrastructure/publishers/{id}` endpoint changed the shape of the `connected_apps` field. It now returns an **array of objects**:

```json
"connected_apps": [
  {
    "access_method": "client",
    "host": "172.20.36.40",
    "last_connected": null,
    "name": "[BD-Gartner-HPE]"
  }
]
```

The SDK struct had `ConnectedApps []string` (array of strings), causing Go's JSON unmarshaler to fail when it tries to decode an object into a string element.

**Note:** The bug only manifests for publishers with at least one connected app. Newly created test publishers with no apps return an empty `connected_apps` array, which explains why acceptance tests did not catch this.

The issue reporter attributed the failure to `upgrade_status` — but `upgrade_status` was already correctly typed as an object in the SDK. The actual culprit was `connected_apps`.

The list endpoint (`GET /infrastructure/publishers`) still returns `connected_apps` as an array of strings. Only the GET-by-ID endpoint changed shape.

## Fix

Changed the `connected_apps` field annotation in `publisher_response.data` from `x-speakeasy-terraform-ignore: true` to `x-speakeasy-ignore: true` in the OAS:

```
/Users/jharris/speakeasy/netskope-apiv2-oas/endpoints/infrastructure/npa_publishers.yaml
```

`x-speakeasy-ignore: true` removes the field entirely from the generated SDK struct, so the JSON deserializer skips it. Since the field was already hidden from Terraform state via `x-speakeasy-terraform-ignore`, this is a no-op from the user's perspective — publishers still work correctly, connected apps are just not surfaced in Terraform state (by design, pre-existing behaviour).

Regenerated SDK via `speakeasy run`.

## Files Changed

- `netskope-apiv2-oas/endpoints/infrastructure/npa_publishers.yaml` — annotation change
- `internal/sdk/models/shared/publisherresponse.go` — `ConnectedApps` field removed
- `internal/provider/npapublisher_resource_sdk.go` — removed `ConnectedApps` mapping

## Tests

- `TestAccNPAPublisher_basic` — passes (was passing before; now covers publisher refresh without error)
- New drift detection test added to verify no perpetual diff for publishers with connected apps

## API Inconsistency

The `GET /infrastructure/publishers` (list) and `GET /infrastructure/publishers/{id}` (get-by-ID) return different shapes for `connected_apps` for the same publisher. The following responses were captured from the bespin tenant against publisher `10950` (`BD-gartner-hpe-sd-wan`).

### `GET /api/v2/infrastructure/publishers` (list)

`connected_apps` is an **array of strings**:

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

### `GET /api/v2/infrastructure/publishers/10950` (get-by-ID)

`connected_apps` is an **array of objects**:

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

The list endpoint also omits fields present in the get-by-ID response (`id`, `labels`, `publisher_upgrade_profiles_id`) and uses different field names for the same values (`publisher_id` vs `id`, `publisher_name` vs `name`, `publisher_upgrade_profiles_external_id` vs `publisher_upgrade_profiles_id`).

Logged in `docs/KNOWN_API_ISSUES.md`.
