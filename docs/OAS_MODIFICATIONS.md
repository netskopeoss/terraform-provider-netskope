# OpenAPI Specification Modifications

This document details the modifications made to the Netskope OpenAPI Specification (OAS) files for use with the Terraform provider. The modifications enable proper Speakeasy SDK code generation and address various API behavioral quirks.

## Overview

**Production OAS Location:** `/path/to/your/terraform-endpoints/endpoints/`
**Modified OAS Location:** `/path/to/your/netskope-apiv2-oas/endpoints/`

The Terraform provider uses modified OAS files that include:
1. Speakeasy-specific annotations for Terraform code generation
2. Schema corrections to match actual API behavior
3. Response wrapper fixes for proper deserialization
4. Field naming consistency improvements

---

## Modified Files

| File | Production Path | Modified Path |
|------|-----------------|---------------|
| Private Apps | `steering/npa_apps_private.yaml` | `steering/npa_apps_private.yaml` |
| NPA Policy Rules | `policy/npa_policy.yaml` | `policy/npa_policy.yaml` |
| Policy Groups | `policy/npa_policygroup.yaml` | `policy/npa_policygroup.yaml` |
| Publishers | `infrastructure/npa_publishers.yaml` | `infrastructure/npa_publishers.yaml` |
| Upgrade Profiles | `infrastructure/npa_upgrade_profiles.yaml` | `infrastructure/npa_upgrade_profiles.yaml` |

---

## Detailed Changes by File

### 1. `npa_apps_private.yaml` (Private Apps)

#### 1.1 Entity Annotations
```yaml
# Added Speakeasy entity annotation for Terraform resource mapping
x-speakeasy-entity: NPAPrivateAppsList
```
**Reason:** Required by Speakeasy to generate proper Terraform data source and resource types.

#### 1.2 Protocol Schema Reference Change
```yaml
# Production
$ref: "#/components/schemas/protocol_response_item"

# Modified
$ref: '#/components/schemas/protocol_item'
```
**Reason:** The API uses `transport` in responses but expects `type` in requests. A unified `protocol_item` schema with proper field mapping resolves this inconsistency. See [KNOWN_API_ISSUES.md Issue #9](./KNOWN_API_ISSUES.md#9-protocol-field-name-mismatch-type-vs-transport).

#### 1.3 Computed Field Annotations
```yaml
# Added to fields like modify_time, modified_by, protocols, public_host
x-speakeasy-computed: true
x-speakeasy-param-suppress-computed-diff: true
```
**Reason:** These fields are set by the API and should not cause Terraform plan diffs. The `suppress-computed-diff` annotation prevents perpetual drift in `terraform plan`.

#### 1.4 Terraform-Ignore Annotations
```yaml
# Added to write-only or response-only fields
x-speakeasy-terraform-ignore: true
```
**Reason:** Fields like `app_option`, `paths`, and `allow_uri_bypass` are not returned by GET operations, causing state drift. Excluding them from the schema prevents issues. See [KNOWN_API_ISSUES.md Issue #10](./KNOWN_API_ISSUES.md#10-write-only-fields-not-returned-in-response).

#### 1.5 Removed Unused Fields
```yaml
# Removed from production schema
hide_app_in_portal:
  type: boolean
custom_host:
  type: string
paths:
  items:
    $ref: "#/components/schemas/path_with_display_name"
  type: array
```
**Reason:** These fields are not consistently returned by the API and cause deserialization issues when present but null/empty.

#### 1.6 Publisher Field Name Override
```yaml
publishers:
  x-speakeasy-name-override: publishers
```
**Reason:** Ensures consistent naming between request and response for the publishers array.

---

### 2. `npa_policy.yaml` (NPA Rules)

#### 2.1 Response Wrapper Structure
```yaml
# Production (incorrect - flat array)
npa_policy_response_list:
  items:
    $ref: "#/components/schemas/npa_policy_response_item"
  type: array

# Modified (correct - wrapped response)
npa_policy_response_list:
  x-speakeasy-entity: NPARulesList
  properties:
    data:
      items:
        $ref: '#/components/schemas/npa_policy_response_item'
      type: array
    status:
      enum:
      - success
      - error
      type: string
  type: object
```
**Reason:** The API returns wrapped responses `{"data": [...], "status": "success"}` but the production OAS incorrectly defined it as a flat array. This caused deserialization failures. See [KNOWN_API_ISSUES.md Issue #12](./KNOWN_API_ISSUES.md#12-npa-rules-response-wrapper-mismatch).

#### 2.2 Response Item Fields Added
```yaml
# Added to npa_policy_response_item schema
enabled:
  example: '1'
  type: string
group_id:
  example: '1'
  type: string
group_name:
  example: My policy group
  type: string
modify_by:
  example: user@example.com
  type: string
modify_time:
  example: '2025-01-01 12:00:00'
  type: string
modify_type:
  example: Created
  type: string
policy_type:
  example: private-app
  type: string
```
**Reason:** The production schema was missing these response fields that are returned by the API.

#### 2.3 ID Field Type and Override
```yaml
# Production
rule_id:
  example: 1
  type: integer

# Modified
rule_id:
  example: '1'
  type: string
  x-speakeasy-name-override: id
```
**Reason:** The API returns rule IDs as strings, not integers. The name override maps to Terraform's standard `id` attribute.

#### 2.4 Default Values for Optional Fields
```yaml
# Added defaults to prevent nil pointer issues
access_method:
  default: []
b_negate_net_location:
  default: false
json_version:
  default: 3
policy_type:
  default: "private-app"
```
**Reason:** Prevents nil/null issues when optional fields are not specified in Terraform configurations.

---

### 3. `npa_policygroup.yaml` (Policy Groups)

#### 3.1 Response Wrapper Structure
```yaml
# Production (incorrect - flat array)
npa_policygroup_response_list:
  items:
    $ref: "#/components/schemas/npa_policygroup_response_item"
  type: array

# Modified (correct - wrapped response)
npa_policygroup_response_list:
  x-speakeasy-entity: NPAPolicyGroupsList
  properties:
    data:
      type: array
      items:
        $ref: '#/components/schemas/npa_policygroup_response_item'
    status:
      type: string
  type: object
```
**Reason:** Same as NPA Rules - the API returns wrapped responses. See [KNOWN_API_ISSUES.md Issue #11](./KNOWN_API_ISSUES.md#11-policy-group-response-wrapper-mismatch).

#### 3.2 Path Correction
```yaml
# Production
/npa/policygroups:

# Modified
/policy/npa/policygroups:
```
**Reason:** The production path was incorrect; the actual API endpoint includes `/policy/` prefix.

#### 3.3 Operation ID and Entity Operation Added
```yaml
# Added for proper Speakeasy mapping
operationId: listNPAPolicyGroups
x-speakeasy-entity-operation: NPAPolicyGroupsList#read
```
**Reason:** Required for Speakeasy to generate proper CRUD operations for Terraform.

#### 3.4 ID Field Types Changed
```yaml
# Production (integer types)
group_id:
  type: integer
id:
  type: integer

# Modified (string types)
group_id:
  type: string
id:
  type: string
  x-speakeasy-name-override: id
```
**Reason:** The API returns these as strings, and Terraform requires string IDs.

#### 3.5 Group Order Schema Flattened
```yaml
# Production (nested object)
group_order:
  properties:
    group_id:
      type: string
    order:
      type: string
  type: object

# Modified (flattened)
group_id:
  example: '1'
  type: string
order:
  enum:
  - before
  - after
  type: string
```
**Reason:** Simplified schema structure for better Terraform resource definition.

---

### 4. `npa_publishers.yaml` (Publishers)

#### 4.1 Protocol Schema Unified
```yaml
# Production
protocol_response_item:
  properties:
    id:
      type: integer
    port:
      type: string
    transport:
      type: string
    # ...

# Modified
protocol_item:
  description: Protocol configuration - type field for tcp/udp value
  properties:
    port:
      type: string
    type:
      x-speakeasy-name-override: protocol
      type: string
      example: tcp
      enum:
      - tcp
      - udp
    # ... with computed annotations
```
**Reason:** Unifies request (`type`) and response (`transport`) field naming. The `x-speakeasy-name-override` annotation handles the mapping.

#### 4.2 Computed Field Annotations
```yaml
# Added to response-only fields
created_at:
  x-speakeasy-computed: true
  x-speakeasy-param-suppress-computed-diff: true
updated_at:
  x-speakeasy-computed: true
  x-speakeasy-param-suppress-computed-diff: true
service_id:
  x-speakeasy-computed: true
  x-speakeasy-param-suppress-computed-diff: true
```
**Reason:** These fields are set by the API and should not trigger plan changes.

---

### 5. `npa_upgrade_profiles.yaml` (Upgrade Profiles)

#### 5.1 Primarily Formatting Changes
Most changes in this file are formatting/style adjustments:
- Quote style normalization (`"` to `'`)
- Array indentation fixes
- Multi-line string formatting

#### 5.2 Schema References Updated
```yaml
# Consistent schema references
$ref: '#/components/schemas/tag_item'
$ref: '#/components/schemas/upgrade_publisher_response'
```
**Reason:** Ensures consistent reference format across the specification.

---

## Overlay File

**File:** `terraform_overlay.yaml`

```yaml
overlay: 1.0.0
info:
  title: Overlay to produce the netskope TF Provider
  version: 0.0.1
actions:
  - target: "info.title"
    update: "Netskope Terraform Provider"
  - target: "info.description"
    update: "Combined specification to produce netskope terraform provider via speakeasy"
  - target: "$.paths.*[?(@.x-netskope-params.customerFacing == false)]"
    remove: true
```

**Purpose:**
1. Updates the API title and description for the Terraform provider
2. Removes non-customer-facing endpoints (internal APIs) from the generated provider

---

## Summary of Key Modifications

| Category | Change Type | Affected Files |
|----------|-------------|----------------|
| Response Wrappers | Schema structure correction | `npa_policy.yaml`, `npa_policygroup.yaml` |
| Field Type Fixes | Integer → String for IDs | `npa_policy.yaml`, `npa_policygroup.yaml` |
| Protocol Schema | Unified request/response naming | `npa_apps_private.yaml`, `npa_publishers.yaml` |
| Computed Fields | Suppress diff annotations | All files |
| Entity Annotations | Speakeasy Terraform mapping | All files |
| Default Values | Prevent nil/null issues | `npa_policy.yaml` |
| Write-Only Fields | Terraform ignore annotations | `npa_apps_private.yaml` |

---

## Related Documentation

- [KNOWN_API_ISSUES.md](./KNOWN_API_ISSUES.md) - Runtime API behavioral issues and workarounds
- [Speakeasy Terraform Documentation](https://www.speakeasy.com/docs/create-terraform) - Speakeasy annotation reference

---

## Maintenance Notes

When updating OAS files:

1. **Always compare with production** before making changes
2. **Test with `speakeasy run`** to verify code generation
3. **Run the CRUD test suite** to validate changes (see separate test repository)
4. **Update this document** with any new modifications
5. **Cross-reference KNOWN_API_ISSUES.md** for related runtime workarounds
