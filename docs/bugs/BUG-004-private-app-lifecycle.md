# netskope_npa_private_app — Config & State Lifecycle

Pseudo-code tracing the full lifecycle of a `netskope_npa_private_app` resource
through Terraform plan, create, read, update, and subsequent plan phases.
Shows exactly where list ordering enters the picture and where drift originates.

---

## 1. TERRAFORM PLAN (first run — no state exists)

```
USER writes main.tf:
  resource "netskope_npa_private_app" "example" {
    private_app_name     = "my-app"
    private_app_hostname = "10.0.0.1"
    protocols = [
      { port = "443", protocol = "tcp" },    # <-- user's chosen order
      { port = "22",  protocol = "tcp" },
    ]
    publishers = [
      { publisher_id = "200", publisher_name = "pub-b" },   # higher ID first
      { publisher_id = "100", publisher_name = "pub-a" },
    ]
    tags = [
      { tag_name = "stage" },       # <-- user's chosen order
      { tag_name = "infosec" },
    ]
  }

TERRAFORM CLI:
  1. Parse HCL config into internal config representation
  2. Load prior state from terraform.tfstate → EMPTY (first run)
  3. Call provider's Schema() to get resource schema definition
     → returns ListNestedAttribute for protocols, publishers, tags
     → all three are Computed: true, Optional: true
  4. Build PLANNED STATE from config:
     planned_state = {
       private_app_name:     "my-app"
       private_app_hostname: "10.0.0.1"
       private_app_id:       <unknown>            # Computed-only, no value yet
       real_host:            <unknown>             # Computed-only
       protocols: [                                # ORDER FROM CONFIG (user's order)
         { port: "443", protocol: "tcp" },
         { port: "22",  protocol: "tcp" },
       ]
       publishers: [                               # ORDER FROM CONFIG
         { publisher_id: "200", publisher_name: "pub-b" },
         { publisher_id: "100", publisher_name: "pub-a" },
       ]
       tags: [                                     # ORDER FROM CONFIG
         { tag_name: "stage",   tag_id: <unknown> },   # tag_id is Computed-only
         { tag_name: "infosec", tag_id: <unknown> },
       ]
     }
  5. Compare planned_state vs prior_state (empty)
     → Everything is new → show "will be created"
  6. *** IF ModifyPlan() exists on the resource, call it here ***
     → Currently: does not exist (no ModifyPlan method)
     → Proposed fix: would run here, but skips because state is null (create)
  7. Display plan to user: "1 resource to add"
```

---

## 2. TERRAFORM APPLY — CREATE

```
TERRAFORM calls provider Create() method
  FILE: npaprivateapp_resource.go:212

  STEP 1 — Extract plan data
    req.Plan.Get(ctx, &plan)           # plan is types.Object
    plan.As(ctx, &data, ...)           # data is *NPAPrivateAppResourceModel
    # data.Protocols = [{Port:"443", Protocol:"tcp"}, {Port:"22", Protocol:"tcp"}]
    # data.Publishers = [{ID:"200", Name:"pub-b"}, {ID:"100", Name:"pub-a"}]
    # data.Tags = [{TagName:"stage"}, {TagName:"infosec"}]
    # (order matches user's HCL)

  STEP 2 — Convert to SDK request object
    FILE: npaprivateapp_resource_sdk.go:301 (ToSharedPrivateAppsRequest)
    request = data.ToSharedPrivateAppsRequest(ctx)
    # Iterates data.Protocols in order → builds []shared.ProtocolItem in same order
    # Iterates data.Publishers in order → builds []shared.PublisherItem in same order
    # Iterates data.Tags in order → builds []shared.TagItemNoID (tag_name only, no tag_id)
    # Order preserved: [tcp:443, tcp:22], [pub-200, pub-100], [stage, infosec]

  STEP 3 — SDK sends HTTP POST to Netskope API
    POST /api/v2/steering/apps/private
    Body: {"app_name":"my-app", "host":"10.0.0.1",
           "protocols":[{"port":"443","type":"tcp"},{"port":"22","type":"tcp"}],
           "publishers":[{"publisher_id":"200"},{"publisher_id":"100"}],
           "tags":[{"tag_name":"stage"},{"tag_name":"infosec"}]}

    HOOK CHAIN — BeforeRequest:
      1. privateAppRequestHook.BeforeRequest()
         FILE: hookPrivateAppRequest.go:23
         → Only fires for "updateNPAPrivateApp", not "createNPAPrivateApps"
         → Passes through unchanged

    API processes request, creates the private app, returns response:
    HTTP 200:
    {"status":"success","data":{
      "app_id":42, "app_name":"[my-app]",
      "protocols":[{"port":"22","transport":"tcp"},{"port":"443","transport":"tcp"}],
                   ^^^^^^^^^ API returns in ITS OWN order (non-deterministic!)
      "service_publisher_assignments":[
        {"publisher_id":100,"publisher_name":" pub-a"},
                                        ^^^ leading whitespace from API!
        {"publisher_id":200,"publisher_name":"pub-b"}
      ],                   ^^^ API returned as integers, not strings!
      "tags":[{"tag_id":7,"tag_name":"infosec"},{"tag_id":3,"tag_name":"stage"}]
    }}         ^^^ API assigned tag_ids; order is by tag_id ascending here (but not guaranteed)

    HOOK CHAIN — AfterSuccess (runs on HTTP response before SDK parses it):
      1. errorStatusResponse.AfterSuccess()
         → Checks for {"status":"error"} with HTTP 200 → not the case → passes through

      2. myAppResponse.AfterSuccess()
         FILE: hookMyAppAfterSuccess.go:27
         Matches operationID "createNPAPrivateApps" → processes response:

         a. Trim app name brackets:
            "[my-app]" → "my-app"

         b. Copy transport → type for each protocol:
            {"port":"22","transport":"tcp"} → {"port":"22","transport":"tcp","type":"tcp"}
            {"port":"443","transport":"tcp"} → {"port":"443","transport":"tcp","type":"tcp"}

         c. SORT protocols by type ASC, then port ASC (numeric):
            BEFORE: [{port:"22",type:"tcp"}, {port:"443",type:"tcp"}]
            AFTER:  [{port:"22",type:"tcp"}, {port:"443",type:"tcp"}]   # already sorted

         d. Convert publisher_id from number to string, trim publisher_name whitespace:
            {publisher_id: 100, publisher_name: " pub-a"} → {publisher_id: "100", publisher_name: "pub-a"}
            {publisher_id: 200, publisher_name: "pub-b"}  → {publisher_id: "200", publisher_name: "pub-b"}

         e. SORT publishers by publisher_id ASC (numeric):
            BEFORE: [{id:"100",name:"pub-a"}, {id:"200",name:"pub-b"}]
            AFTER:  [{id:"100",name:"pub-a"}, {id:"200",name:"pub-b"}]   # already sorted

         f. SORT tags by tag_id ASC:
            BEFORE: [{tag_id:7,tag_name:"infosec"}, {tag_id:3,tag_name:"stage"}]
            AFTER:  [{tag_id:3,tag_name:"stage"}, {tag_id:7,tag_name:"infosec"}]
                     ^^^^^^^^ REORDERED by tag_id

         → Marshals modified response back into HTTP response body

      3. myBulkAppResponse.AfterSuccess()
         → operationID "createNPAPrivateApps" ≠ "listNPAPrivateApps" → passes through

  STEP 4 — SDK parses modified HTTP response into shared.PrivateAppsPostResponse
    SDK unmarshals JSON → shared.PrivateAppsPostResponseData struct
    Lists are in hook-sorted order now.

  STEP 5 — Refresh model from POST response
    FILE: npaprivateapp_resource.go:256
    data.RefreshFromSharedPrivateAppsPostResponseData(ctx, res.Data)
    FILE: npaprivateapp_resource_sdk.go:83
    # Iterates resp.Protocols in hook-sorted order → populates data.Protocols
    # Note: POST response does NOT include publishers, so data.Publishers unchanged
    # Iterates resp.Tags in hook-sorted order → populates data.Tags

  STEP 6 — refreshPlan() merges plan unknowns into data
    FILE: utils.go:76
    # Handles Computed+Optional attributes: if plan had unknown, keep state value
    # This resolves <unknown> for private_app_id, real_host, etc.

  STEP 7 — Follow-up GET to read full resource
    FILE: npaprivateapp_resource.go:273
    # Calls GetNPAPrivateApp to get complete data (POST response may omit fields)
    # Same hook chain runs on GET response:
    #   → errorStatusResponse → passes through
    #   → myAppResponse (operationID "getNPAPrivateApp") → sorts again
    #   → myBulkAppResponse → passes through

  STEP 8 — Refresh model from GET response
    FILE: npaprivateapp_resource.go:293
    data.RefreshFromSharedPrivateAppsItem(ctx, res.Data)
    FILE: npaprivateapp_resource_sdk.go:15
    # Iterates all fields in hook-sorted order → populates data

    data is now:
      Protocols:  [{Port:"22", Protocol:"tcp"}, {Port:"443", Protocol:"tcp"}]    # sorted
      Publishers: [{ID:"100", Name:"pub-a"}, {ID:"200", Name:"pub-b"}]           # sorted
      Tags:       [{TagID:3, TagName:"stage"}, {TagID:7, TagName:"infosec"}]     # sorted

  STEP 9 — refreshPlan() again
    # Merges any remaining plan values

  STEP 10 — Write to state
    FILE: npaprivateapp_resource.go:306
    resp.State.Set(ctx, &data)
    # Terraform serializes data into terraform.tfstate

    STATE IS NOW (terraform.tfstate):
      protocols: [
        { port: "22",  protocol: "tcp" },     ← hook-sorted (port ASC)
        { port: "443", protocol: "tcp" },
      ]
      publishers: [
        { publisher_id: "100", publisher_name: "pub-a" },   ← hook-sorted (id ASC)
        { publisher_id: "200", publisher_name: "pub-b" },
      ]
      tags: [
        { tag_id: 3, tag_name: "stage" },      ← hook-sorted (tag_id ASC)
        { tag_id: 7, tag_name: "infosec" },
      ]

    NOTE: State order ≠ config order for publishers and tags!
    Config had: publishers [200, 100], tags [stage, infosec]
    State has:  publishers [100, 200], tags [stage, infosec] (happened to match for tags here)
```

---

## 3. TERRAFORM PLAN (second run — state exists, NO config changes)

**This is where BUG-002 manifests.**

```
TERRAFORM CLI:
  1. Parse HCL config (unchanged from first run)
     config_protocols:  [{port:"443", protocol:"tcp"}, {port:"22",  protocol:"tcp"}]
     config_publishers: [{id:"200", name:"pub-b"}, {id:"100", name:"pub-a"}]
     config_tags:       [{tag_name:"stage"}, {tag_name:"infosec"}]

  2. Load prior state from terraform.tfstate
     state_protocols:   [{port:"22",  protocol:"tcp"}, {port:"443", protocol:"tcp"}]  # sorted
     state_publishers:  [{id:"100", name:"pub-a"}, {id:"200", name:"pub-b"}]          # sorted
     state_tags:        [{tag_id:3, tag_name:"stage"}, {tag_id:7, tag_name:"infosec"}] # sorted

  3. Build PLANNED STATE from config + state:
     For Computed+Optional list attributes where config provides a value:
       → Plan uses CONFIG order for the list
       → Computed-only sub-attributes (tag_id) are marked <unknown>

     planned_state = {
       protocols: [
         { port: "443", protocol: "tcp" },     ← FROM CONFIG (user order)
         { port: "22",  protocol: "tcp" },
       ]
       publishers: [
         { publisher_id: "200", publisher_name: "pub-b" },  ← FROM CONFIG
         { publisher_id: "100", publisher_name: "pub-a" },
       ]
       tags: [
         { tag_name: "stage",   tag_id: <unknown> },   ← FROM CONFIG (tag_id is Computed-only)
         { tag_name: "infosec", tag_id: <unknown> },
       ]
     }

  4. *** ModifyPlan() WOULD run here (proposed fix) ***
     Currently: does not exist → plan is unchanged
     Proposed: normalizes plan order to match state → no diff

  5. COMPARE planned_state vs prior_state (POSITION BY POSITION):

     protocols[0]: plan={port:"443"} vs state={port:"22"}   → DIFFERENT! (port changed)
     protocols[1]: plan={port:"22"}  vs state={port:"443"}  → DIFFERENT! (port changed)
     *** DRIFT DETECTED *** (false positive — same elements, different order)

     publishers[0]: plan={id:"200"} vs state={id:"100"}     → DIFFERENT!
     publishers[1]: plan={id:"100"} vs state={id:"200"}     → DIFFERENT!
     *** DRIFT DETECTED *** (false positive)

     tags[0]: plan={name:"stage", tag_id:<unknown>} vs state={name:"stage", tag_id:3}
              → tag_id: 3 → <unknown>                       → DIFFERENT!
     tags[1]: plan={name:"infosec", tag_id:<unknown>} vs state={name:"infosec", tag_id:7}
              → tag_id: 7 → <unknown>                       → DIFFERENT!
     *** DRIFT DETECTED *** (false positive — computed value showing as "known after apply")

  6. Terraform shows:
     ~ resource "netskope_npa_private_app" "example" {
         ~ protocols = [
             ~ { ~ port = "22" -> "443" },
             ~ { ~ port = "443" -> "22" },
           ]
         ~ publishers = [
             ~ { ~ publisher_id = "100" -> "200", ~ publisher_name = "pub-a" -> "pub-b" },
             ~ { ~ publisher_id = "200" -> "100", ~ publisher_name = "pub-b" -> "pub-a" },
           ]
         ~ tags = [
             ~ { tag_name = "stage",   ~ tag_id = 3 -> (known after apply) },
             ~ { tag_name = "infosec", ~ tag_id = 7 -> (known after apply) },
           ]
       }

     "1 resource to update" — EVERY RUN, FOREVER
```

---

## 4. TERRAFORM APPLY — UPDATE (triggered by false drift)

```
TERRAFORM calls provider Update() method
  FILE: npaprivateapp_resource.go:367

  STEP 1 — Merge plan + state into data
    FILE: utils.go:48 (merge function)
    merge(ctx, req, resp, &data)
    # Starts from STATE, overlays PLAN values
    # data.Protocols now has CONFIG order: [{tcp:443}, {tcp:22}]
    # data.Publishers now has CONFIG order: [{200,pub-b}, {100,pub-a}]
    # data.Tags now has CONFIG order: [{stage}, {infosec}]

  STEP 2 — Convert to SDK update request
    FILE: npaprivateapp_resource_sdk.go:162 (ToOperationsUpdateNPAPrivateAppRequest)
    → calls ToSharedPrivateAppsPutRequest (line 183)
    # Iterates data in config order → builds request in config order

  STEP 3 — SDK sends HTTP PUT to Netskope API
    PUT /api/v2/steering/apps/private/42
    Body: same elements as before, but in config order

    HOOK CHAIN — BeforeRequest:
      1. privateAppRequestHook.BeforeRequest()
         FILE: hookPrivateAppRequest.go:23
         → operationID "updateNPAPrivateApp" matches!
         → Removes empty "paths", "app_option", "uribypass_header_value" if present
         → Does NOT reorder lists (just cleans empty fields)

    API processes PUT → returns updated resource
    (API reorders elements internally; response order is non-deterministic)

    HOOK CHAIN — AfterSuccess:
      → myAppResponse.AfterSuccess() sorts protocols, publishers, tags
      → Same sorting as create: type/port, publisher_id, tag_id

  STEP 4-9 — Same as Create steps 5-10
    Refresh from PUT response → refreshPlan → follow-up GET → refresh again → write state

    STATE IS NOW: same sorted order as before (hooks ensure determinism)
      protocols:  [{tcp:22}, {tcp:443}]        # sorted
      publishers: [{100,pub-a}, {200,pub-b}]   # sorted
      tags:       [{tag_id:3,stage}, {tag_id:7,infosec}]  # sorted

  RESULT: Unnecessary API call was made. Nothing actually changed.
          Next plan will show the SAME drift again → infinite loop.
```

---

## 5. TERRAFORM PLAN — READ (refresh)

```
When terraform runs "plan" or "apply", it calls Read() first to refresh state.

TERRAFORM calls provider Read() method
  FILE: npaprivateapp_resource.go:309

  STEP 1 — Load current state
    req.State.Get(ctx, &item) → item.As(ctx, &data, ...)
    # data loaded from terraform.tfstate (hook-sorted order from last apply)

  STEP 2 — Build GET request
    data.ToOperationsGetNPAPrivateAppRequest(ctx)
    → operations.GetNPAPrivateAppRequest{PrivateAppID: 42}

  STEP 3 — SDK sends HTTP GET
    GET /api/v2/steering/apps/private/42

    API returns response (possibly different order than last time!)

    HOOK CHAIN — AfterSuccess:
      → myAppResponse.AfterSuccess() (operationID "getNPAPrivateApp")
        - Trims app name brackets
        - Copies transport → type
        - SORTS protocols by type/port
        - Converts publisher_id to string, trims whitespace
        - SORTS publishers by publisher_id
        - SORTS tags by tag_id
      → Response is now in deterministic order regardless of API's internal ordering

  STEP 4 — Refresh model
    data.RefreshFromSharedPrivateAppsItem(ctx, res.Data)
    # Populates data in hook-sorted order

  STEP 5 — Write refreshed state
    resp.State.Set(ctx, &data)
    # State is updated with hook-sorted order (same as before if nothing changed on API side)

  RESULT: State is always deterministic after Read.
          The problem is NOT in Read — it's in the plan comparison (step 3 above).
```

---

## 6. PROPOSED FIX — ModifyPlan lifecycle

```
WHERE IT FITS (between steps 3 and 5 of "TERRAFORM PLAN"):

  TERRAFORM CLI:
    1. Parse HCL config
    2. Load prior state
    3. Build planned_state from config
       planned_state.protocols = [{tcp:443}, {tcp:22}]     # config order
       planned_state.tags = [{stage,<unknown>}, {infosec,<unknown>}]

    4. *** Call provider's ModifyPlan() ***
       FILE: npaprivateapp_resource_planmodify.go (NEW)

       ModifyPlan(ctx, req, resp):
         req.State.Raw.IsNull()?  → No (state exists)
         req.Plan.Raw.IsNull()?   → No (not a destroy)

         normalizeProtocolsOrder(ctx, req, resp):
           plan_protocols  = [{tcp:443}, {tcp:22}]          from req.Plan
           state_protocols = [{tcp:22}, {tcp:443}]          from req.State

           len(plan) == len(state)?  → Yes (2 == 2)
           any unknown fingerprint keys?  → No (port and protocol both known)

           plan fingerprints:  {"tcp:443": 1, "tcp:22": 1}
           state fingerprints: {"tcp:22": 1, "tcp:443": 1}
           match?  → YES (same multiset)

           → resp.Plan.SetAttribute("protocols", state_protocols)
           → Plan protocols is now [{tcp:22}, {tcp:443}]    # matches state!

         normalizePublishersOrder(ctx, req, resp):
           plan_publishers  = [{id:"200"}, {id:"100"}]
           state_publishers = [{id:"100"}, {id:"200"}]

           plan fingerprints:  {"200": 1, "100": 1}
           state fingerprints: {"100": 1, "200": 1}
           match?  → YES

           → resp.Plan.SetAttribute("publishers", state_publishers)
           → Plan publishers is now [{id:"100"}, {id:"200"}]  # matches state!

         normalizeTagsOrder(ctx, req, resp):
           plan_tags  = [{name:"stage",id:<unknown>}, {name:"infosec",id:<unknown>}]
           state_tags = [{name:"stage",id:3}, {name:"infosec",id:7}]

           plan fingerprints (by tag_name):  {"stage": 1, "infosec": 1}
           state fingerprints:               {"stage": 1, "infosec": 1}
           match?  → YES

           → resp.Plan.SetAttribute("tags", state_tags)
           → Plan tags is now [{name:"stage",id:3}, {name:"infosec",id:7}]
             ^^^ Uses STATE values including tag_id — no more "(known after apply)"!

    5. COMPARE modified planned_state vs prior_state:
       protocols[0]: {tcp:22} vs {tcp:22}   → SAME
       protocols[1]: {tcp:443} vs {tcp:443} → SAME
       publishers[0]: {100} vs {100}        → SAME
       publishers[1]: {200} vs {200}        → SAME
       tags[0]: {stage,3} vs {stage,3}      → SAME
       tags[1]: {infosec,7} vs {infosec,7}  → SAME

    6. "No changes. Your infrastructure matches the configuration."
```

---

## 7. WHAT HAPPENS WHEN THERE IS A REAL CHANGE

```
USER edits main.tf:
  protocols = [
    { port = "443", protocol = "tcp" },
    { port = "8080", protocol = "tcp" },    # CHANGED: was "22"
  ]

TERRAFORM PLAN:
  planned_state.protocols = [{tcp:443}, {tcp:8080}]

  ModifyPlan() runs:
    normalizeProtocolsOrder:
      plan fingerprints:  {"tcp:443": 1, "tcp:8080": 1}
      state fingerprints: {"tcp:22": 1, "tcp:443": 1}
      match?  → NO ("tcp:8080" not in state, "tcp:22" not in plan)
      → SKIP normalization — don't modify plan

  COMPARE: planned vs state shows real change:
    protocols[0]: {tcp:443} vs {tcp:22}     → CHANGED (real)
    protocols[1]: {tcp:8080} vs {tcp:443}   → CHANGED (real)

  → Terraform correctly shows "1 resource to update"
  → Apply sends PUT with new protocols
  → Hooks sort response → state updated with [{tcp:443}, {tcp:8080}] (sorted)
```

---

## Summary: Data Flow Diagram

```
                    ┌─────────────┐
                    │   HCL Config │  (user's .tf files)
                    │   ORDER: A,B │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Terraform   │  Parses config
                    │  Plan Phase  │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Build Plan   │  planned_state = config values
                    │ ORDER: A,B   │  (config order preserved)
                    └──────┬──────┘
                           │
              ┌────────────▼────────────┐
              │ ModifyPlan() [PROPOSED] │  if same set → use state order
              │ ORDER: B,A (from state) │
              └────────────┬────────────┘
                           │
                    ┌──────▼──────┐
                    │  Compare     │  plan vs state
                    │  B,A vs B,A  │  → NO DIFF ✓
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  (if diff)   │  Call Create/Update
                    │  Provider    │
                    └──────┬──────┘
                           │
               ┌───────────▼───────────┐
               │ ToSDK conversion      │  NPAPrivateAppResourceModel → shared.PrivateAppsRequest
               │ (preserves order)     │  npaprivateapp_resource_sdk.go
               └───────────┬───────────┘
                           │
               ┌───────────▼───────────┐
               │ SDK HTTP Client       │
               │ sends POST/PUT/GET    │
               └───────────┬───────────┘
                           │
            ┌──────────────▼──────────────┐
            │ BeforeRequest Hooks         │
            │ 1. privateAppRequestHook    │  strips empty fields (update only)
            │    (hookPrivateAppRequest)  │
            └──────────────┬──────────────┘
                           │
                    ┌──────▼──────┐
                    │ Netskope API │  processes request
                    │  returns     │  response with NON-DETERMINISTIC order
                    └──────┬──────┘
                           │
            ┌──────────────▼──────────────┐
            │ AfterSuccess Hooks          │
            │ 1. errorStatusResponse      │  catches 200+error
            │ 2. myAppResponse            │  ← SORTS protocols, publishers, tags
            │    (hookMyAppAfterSuccess)  │     trims names, converts types
            │ 3. myBulkAppResponse        │  (for list operations)
            │ ORDER: B,A (sorted)         │
            └──────────────┬──────────────┘
                           │
               ┌───────────▼───────────┐
               │ SDK unmarshals JSON   │  → shared.PrivateAppsItem
               │ ORDER: B,A (sorted)   │
               └───────────┬───────────┘
                           │
               ┌───────────▼───────────┐
               │ FromSDK conversion    │  RefreshFromSharedPrivateAppsItem
               │ (preserves order)     │  npaprivateapp_resource_sdk.go
               │ ORDER: B,A (sorted)   │
               └───────────┬───────────┘
                           │
               ┌───────────▼───────────┐
               │ resp.State.Set(data)  │  Write to terraform.tfstate
               │ ORDER: B,A (sorted)   │
               └───────────┘

NEXT PLAN:
  Config order: A,B    (user never changed their .tf)
  State order:  B,A    (sorted by hooks)
  Without ModifyPlan:  A,B ≠ B,A → DRIFT (bug)
  With ModifyPlan:     B,A == B,A → no diff (fix)
```

---

## Key Source Files (in execution order)

| Step | File | Function |
|------|------|----------|
| Schema | `internal/provider/npaprivateapp_resource.go:61` | `Schema()` — defines ListNestedAttribute |
| Plan modify | `internal/provider/npaprivateapp_resource_planmodify.go` | `ModifyPlan()` — **PROPOSED** |
| Create | `internal/provider/npaprivateapp_resource.go:212` | `Create()` |
| Read | `internal/provider/npaprivateapp_resource.go:309` | `Read()` |
| Update | `internal/provider/npaprivateapp_resource.go:367` | `Update()` |
| ToSDK | `internal/provider/npaprivateapp_resource_sdk.go:301` | `ToSharedPrivateAppsRequest()` |
| ToSDK (PUT) | `internal/provider/npaprivateapp_resource_sdk.go:183` | `ToSharedPrivateAppsPutRequest()` |
| FromSDK | `internal/provider/npaprivateapp_resource_sdk.go:15` | `RefreshFromSharedPrivateAppsItem()` |
| Plan merge | `internal/provider/utils.go:76` | `refreshPlan()` |
| State merge | `internal/provider/utils.go:48` | `merge()` |
| Before hook | `internal/sdk/internal/hooks/hookPrivateAppRequest.go:23` | `BeforeRequest()` — strips empties |
| After hook | `internal/sdk/internal/hooks/hookMyAppAfterSuccess.go:27` | `AfterSuccess()` — **SORTS lists** |
| After hook (bulk) | `internal/sdk/internal/hooks/hookMyBulkAppAfterSuccess.go` | `AfterSuccess()` — sorts in bulk |
| Hook registration | `internal/sdk/internal/hooks/registration.go:11` | `initHooks()` |
| Hook framework | `internal/sdk/internal/hooks/hooks.go:133` | `AfterSuccess()` — chains hooks |
