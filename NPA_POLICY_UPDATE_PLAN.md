# Plan: NPA Policy OAS Schema Update with Terraform Annotations

## Objective
Update the NPA policy OAS file with the new request/response schema split AND add all required x-speakeasy annotations for Terraform provider generation.

---

## SIGNIFICANT CHANGES (Action Required)

### 1. TERRAFORM-IGNORE ANNOTATIONS (Main Change)
Add `x-speakeasy-terraform-ignore: true` annotations to response-only fields in existing `npa_policy_rule_data` schema. **No schema split required.**

This completely hides the fields from Terraform schema (simpler approach than `x-speakeasy-param-computed`).

### 2. FIELDS REQUIRING TERRAFORM-IGNORE
The following fields in `npa_policy_rule_data` need terraform-ignore annotations:

| Field | Annotation | Purpose |
|-------|------------|---------|
| `dlp_actions` | `x-speakeasy-terraform-ignore: true` | DLP config set by backend |
| `tss_actions` | `x-speakeasy-terraform-ignore: true` | TSS config set by backend |
| `tss_profile` | `x-speakeasy-terraform-ignore: true` | TSS profiles set by backend |
| `external_dlp` | `x-speakeasy-terraform-ignore: true` | Flag set by backend |
| `privateAppsWithActivities` | `x-speakeasy-terraform-ignore: true` | Computed app activities |
| `show_dlp_profile_action_table` | `x-speakeasy-terraform-ignore: true` | UI flag set by backend |
| `schedule` | `x-speakeasy-terraform-ignore: true` | **NEW** - Schedule config |

**Why:** These fields are API-computed and not user-settable. Hiding them from Terraform prevents drift.

### 3. NEW SCHEMA: `npa_schedule`
New schema for policy scheduling support. This is a response-only computed field.

### 4. PATHS (No Change)
Paths remain unchanged:
- `/policy/npa/rules`
- `/policy/npa/rules/{id}`

### 5. REFERENCE FILE
A complete suggested OAS file with all annotations has been generated:
- `npa_policy_suggested.yaml` (in this project directory)

---

## Summary of Changes

**Schema Changes:**
- Add `x-speakeasy-terraform-ignore: true` annotations to response-only fields in `npa_policy_rule_data`
- New `npa_schedule` schema for policy scheduling (hidden from Terraform)
- Response-only fields: `dlp_actions`, `tss_actions`, `tss_profile`, `external_dlp`, `privateAppsWithActivities`, `show_dlp_profile_action_table`, `schedule`
- Paths remain at `/policy/npa/rules`

**Annotations to Add:**
- All x-speakeasy entity/operation annotations
- Terraform-ignore annotations for response-only fields
- Default values for request schema fields

---

## Complete Change Analysis

### 1. Terraform-Ignore Annotations
Add to existing `npa_policy_rule_data` schema (no split required):
- `x-speakeasy-terraform-ignore: true` - completely hides field from Terraform schema

### 2. New Schema Added
- `npa_schedule` - scheduling support for policies (response-only)

### 3. Response-Only Fields
These are in response but NOT in request (computed by API):
- `dlp_actions` - DLP configuration
- `tss_actions` - TSS configuration
- `tss_profile` - TSS profile array
- `external_dlp` - External DLP flag
- `privateAppsWithActivities` - Apps with activities
- `show_dlp_profile_action_table` - UI flag
- `schedule` - Schedule configuration (NEW)

### 4. Response Wrapper Changed
**Old:** `npa_policy_response` was an object with `data` array and `status`
**New:** `npa_policy_response` is a direct array

### 5. Paths (No Change)
Paths remain at `/policy/npa/rules` and `/policy/npa/rules/{id}`

### 6. Required x-speakeasy Annotations

**On schemas:**
```yaml
npa_policy_request:
  x-speakeasy-entity: NPARules
  # ... rest of schema

npa_policy_response:
  x-speakeasy-entity: NPARulesList
  # ... rest of schema

npa_policy_response_item:
  x-speakeasy-entity: NPARules
  properties:
    rule_id:
      x-speakeasy-name-override: id
      # ... rest of field
```

**On response-only fields in `npa_policy_rule_data`:**
```yaml
dlp_actions:
  x-speakeasy-terraform-ignore: true
  $ref: "#/components/schemas/npa_policy_rule_dlp"
tss_actions:
  x-speakeasy-terraform-ignore: true
  $ref: "#/components/schemas/npa_policy_rule_tss"
tss_profile:
  x-speakeasy-terraform-ignore: true
  type: array
  items:
    type: string
external_dlp:
  x-speakeasy-terraform-ignore: true
  type: boolean
privateAppsWithActivities:
  x-speakeasy-terraform-ignore: true
  # ... rest of field
show_dlp_profile_action_table:
  x-speakeasy-terraform-ignore: true
  type: boolean
schedule:
  x-speakeasy-terraform-ignore: true
  $ref: "#/components/schemas/npa_schedule"
```

**On paths:**
```yaml
/policy/npa/rules:
  get:
    operationId: listNPARules
    x-speakeasy-entity-operation: NPARulesList#read
  post:
    operationId: createNPARules
    x-speakeasy-entity-operation: NPARules#create
    parameters:
      - name: silent
        x-speakeasy-ignore: true

/policy/npa/rules/{id}:
  get:
    operationId: getNPARules
    x-speakeasy-entity-operation: NPARules#read
    parameters:
      - name: fields
        x-speakeasy-ignore: true
  patch:
    operationId: updateNPARules
    x-speakeasy-entity-operation: NPARules#update
    parameters:
      - name: silent
        x-speakeasy-ignore: true
  delete:
    operationId: deleteNPARules
    x-speakeasy-entity-operation: NPARules#delete
```

**On response status fields:**
```yaml
# In each response that has a status wrapper:
status:
  x-speakeasy-terraform-ignore: true
```

### 7. Default Values for Request Schema

Add defaults to `npa_policy_request_rule_data`:
```yaml
access_method:
  type: array
  default: []
b_negateNetLocation:
  type: boolean
  default: false
b_negateSrcCountries:
  type: boolean
  default: false
json_version:
  type: integer
  default: 3
device_classification_id:
  type: array
  default: []
net_location_obj:
  type: array
  default: []
organization_units:
  type: array
  default: []
policy_type:
  type: string
  default: "private-app"
privateAppTagIds:
  type: array
  default: []
privateAppTags:
  type: array
  default: []
privateApps:
  type: array
  default: []
srcCountries:
  type: array
  default: []
userGroups:
  type: array
  default: []
users:
  type: array
  default: []
```

---

## Files Affected

### OAS Files (Source)
- `/Users/jharris/speakeasy/netskope-apiv2-oas/endpoints/policy/npa_policy.yaml` - **PRIMARY CHANGE**

### Generated Files (Auto-regenerated)
After running `speakeasy run`, these will be regenerated:

**Provider Files:**
- `internal/provider/nparules_resource.go`
- `internal/provider/nparules_resource_sdk.go`
- `internal/provider/nparules_data_source.go`
- `internal/provider/nparules_data_source_sdk.go`

**Type Files:**
- `internal/provider/types/npa_policy_rule_data.go` → split into two files
- `internal/provider/types/npa_policy_request_rule_data.go` (new)
- `internal/provider/types/npa_policy_response_rule_data.go` (new)

**SDK Models:**
- `internal/sdk/models/shared/npa_policy_*.go`

### Config Files (May need updates)
- `.speakeasy/workflow.yaml` - No change needed
- `terraform_overlay.yaml` - No change needed

---

## Implementation Steps

### Step 1: Update OAS Schema

1. Add `x-speakeasy-terraform-ignore: true` to response-only fields in `npa_policy_rule_data`
2. Add `npa_schedule` schema for scheduling support
3. Add `schedule` field to `npa_policy_rule_data` with terraform-ignore annotation
4. Ensure default values are set on user-settable fields

### Step 2: Verify Path Annotations

1. Paths remain at `/policy/npa/rules` and `/policy/npa/rules/{id}`
2. Ensure all x-speakeasy-entity-operation annotations are present

### Step 3: Regenerate Provider
```bash
cd /Users/jharris/PycharmProjects/terraform-provider-netskope
speakeasy run
```

### Step 4: Test Changes
```bash
# Build provider
go build ./...

# Run in debug mode
go run main.go -debug

# Test with existing terraform configs
cd /Users/jharris/PycharmProjects/terraform-provider-netskope-tests/crud
terraform plan
terraform apply
terraform destroy
```

---

## Verification Checklist

- [ ] OAS schema validates (`speakeasy validate`)
- [ ] Provider builds (`go build ./...`)
- [ ] Terraform plan shows no unexpected drift
- [ ] Create/Read/Update/Delete operations work
- [ ] Hidden fields (terraform-ignore) don't appear in Terraform schema
- [ ] No "(known after apply)" for fields with defaults

---

## Questions to Resolve

1. **Should we add `x-speakeasy-entity-version: 2`?**
   - This enables state migration for breaking schema changes
   - Recommended since rule_data structure is changing

2. **Response wrapper - any changes needed?**
   - Current: `{data: [...], status: "success"}`
   - Confirm this remains unchanged

---

## Current State Analysis

The current `npa_policy.yaml` file already has:
- x-speakeasy-entity annotations on schemas (NPARules, NPARulesList)
- x-speakeasy-entity-operation annotations on all paths
- x-speakeasy-ignore on silent and fields parameters
- x-speakeasy-terraform-ignore on status fields
- Default values on many fields

What still needs to be done:
1. Add `x-speakeasy-terraform-ignore: true` annotations to response-only fields in `npa_policy_rule_data`
2. Add `npa_schedule` schema
3. Add `schedule` field with terraform-ignore annotation

---

## Quick Reference: Speakeasy Annotation Guide

| Annotation | Location | Purpose |
|------------|----------|---------|
| `x-speakeasy-entity: NPARules` | Schema | Identifies schema as Terraform resource |
| `x-speakeasy-entity-operation: NPARules#create` | Path operation | Maps API operation to Terraform CRUD |
| `x-speakeasy-name-override: id` | Property | Renames field in Terraform (rule_id → id) |
| `x-speakeasy-ignore: true` | Parameter | Hides parameter from Terraform schema |
| `x-speakeasy-terraform-ignore: true` | Property | **Completely hides field from Terraform schema** |

---

## Files in This Update

| File | Description |
|------|-------------|
| `NPA_POLICY_UPDATE_PLAN.md` | This plan document |
| `npa_policy_suggested.yaml` | Complete suggested OAS with all annotations |
| `npa_policy.yaml` (source) | Original file to be updated |
