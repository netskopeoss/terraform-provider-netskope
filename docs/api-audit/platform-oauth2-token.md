# API Audit: POST /platform/oauth2/token

**Date:** 2026-07-01
**OAS source:** `/Users/jharris/PycharmProjects/api-gateway-endpoints-master/production/endpoints/platform/ms-platform.yaml`
**Terraform resource:** `data.netskope_platform_oauth2_token`

---

## Endpoint

`POST /api/v2/platform/oauth2/token`

RFC 6749 compliant client_credentials grant. No authentication header required (`skipAuthentication: true`).

---

## Request fields

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `client_id` | string | Yes (in practice) | OAuth2 client ID. Required by API even though OAS marks optional — `Unknown or missing client_id` if absent. |
| `client_secret` | string | Yes (in practice) | OAuth2 client secret. Sensitive. |
| `grant_type` | string | Yes | Only `client_credentials` supported. OAS marks required. |
| `scope` | string | No | Accepted as passthrough — API does not act on it. Hidden from Terraform schema (`x-speakeasy-terraform-ignore`). |

**Content-Type:** API accepts both `application/json` and `application/x-www-form-urlencoded`. Terraform implementation uses JSON.

## Response fields

| Field | Type | Notes |
|-------|------|-------|
| `access_token` | string | Bearer token. Short-lived. Sensitive. |
| `token_type` | string | Always `"Bearer"`. |
| `expires_in` | integer | Seconds until expiry (e.g. 3600). |

## Error shape (all error responses)

```json
{
  "error": "invalid_request",
  "error_description": "Unknown or missing client_id"
}
```

RFC 6749 §5.2 format. Tested error codes: `invalid_request`, `invalid_client`, `unsupported_grant_type`.

---

## Verified behaviors

| Behavior | Verified |
|----------|----------|
| Empty body → `grant_type is required` (400) | Yes |
| Missing client_id → `Unknown or missing client_id` (400) | Yes |
| Bad credentials → `invalid_request` not `invalid_client` (400, not 401) | Yes — API returns 400 even for bad creds; OAS says 401 for `invalid_client` |
| `scope` field accepted but ignored | Inferred from OAS description; not tested |

## Discrepancies vs OAS

- **`client_id` is effectively required** even though OAS marks it optional. Omitting it returns `400 invalid_request`.
- **Bad credentials return 400, not 401.** OAS documents `invalid_client` as a 401, but the bespin tenant returned 400 for unknown `client_id`. May differ for known-but-wrong credentials.

---

## Terraform model

**Data source** — `data.netskope_platform_oauth2_token`

```hcl
data "netskope_platform_oauth2_token" "token" {
  client_id     = var.oauth2_client_id
  client_secret = var.oauth2_client_secret  # sensitive
}

output "access_token" {
  value     = data.netskope_platform_oauth2_token.token.access_token
  sensitive = true
}
```

Re-evaluated on every `terraform plan`, so the token is always fresh.

## Related endpoint

`POST /api/v2/platform/oauth2/token/generate` — older proprietary format using `clientID`/`secretKey`. Marked `x-speakeasy-ignore: true` in the OAS; not surfaced in the Terraform provider.
