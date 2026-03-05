# Socket LAN Resources - API Errors Log

This document tracks API and state consistency errors encountered during the development of Socket LAN resources (`cato_socket_lan_section`, `cato_socket_lan_network_rule`, `cato_socket_lan_firewall_rule`).

---

## Error 1: Service Field State Inconsistency (Empty List vs Null)

**Date:** 2026-02-20

**Error Message:**
```
Error: Provider produced inconsistent result after apply
.rule.service.custom: was cty.ListValEmpty(cty.Object(...)), but now null.
```

**Cause:**
User specified `custom = []` (explicit empty list) in Terraform config, but state hydration returned `types.ListNull(...)`. Terraform treats empty list and null as different values.

**Root Cause Analysis:**
- Terraform config: `service = { simple = [...], custom = [] }`
- Plan value: `cty.ListValEmpty(...)`
- API returns: nothing for empty custom services
- State hydration returned: `types.ListNull(...)`
- Result: Plan != State

**Solution:**
1. Added `listvalidator.SizeAtLeast(1)` to `custom` field to prevent users from specifying empty lists
2. Removed `Computed: true` from `service`, `simple`, and `custom` schema attributes (matching IFW pattern)
3. Updated test configs to omit fields instead of specifying empty lists

**Files Modified:**
- `internal/provider/resource_socket_lan_network_rule.go` - Schema changes
- Test config files - Removed `custom = []` patterns

**Key Lesson:**
In Terraform, `field = []` (empty list) is NOT the same as omitting the field (null). If a field is optional and you don't want it, omit it entirely.

---

## Error 2: Service Field State Inconsistency (Null vs Object)

**Date:** 2026-02-20

**Error Message:**
```
Error: Provider produced inconsistent result after apply
.rule.service: was null, but now cty.ObjectVal(map[string]cty.Value{"custom":cty.NullVal(...), "simple":cty.NullVal(...)}).
```

**Cause:**
User did not specify `service` block at all, but state hydration always created a service object with null values inside.

**Root Cause Analysis:**
- Terraform config: No `service` block specified
- Plan value: `null`
- State hydration: Always called `hydrateSocketLanServiceState()` which returned an object
- Result: Plan (null) != State (object with null fields)

**Solution:**
Modified state hydration to only create service object when there's actual content (following IFW pattern):

```go
// Before (wrong):
result.Service = hydrateSocketLanServiceState(ctx, apiRule.Service)

// After (correct):
if len(apiRule.Service.Simple) > 0 || len(apiRule.Service.Custom) > 0 {
    result.Service = hydrateSocketLanServiceState(ctx, apiRule.Service)
} else {
    result.Service = types.ObjectNull(SocketLanServiceAttrTypes)
}
```

**Files Modified:**
- `internal/provider/hydrate_socket_lan_network_rule_state.go`

**Key Lesson:**
Only create optional nested objects in state when they contain actual data. If API returns empty, state should be null to match a config that omits the field.

---

## Error 3: GraphQL Validation - Source VLAN Cannot Be Null

**Date:** 2026-02-20

**Error Message:**
```
Error: Catov2 API PolicySocketLanFirewallAddRule error

{"networkErrors":{"code":422,"message":"Response body {\"errors\":[{\"message\":\"cannot be null\",\"path\":[\"variable\",\"socketLanFirewallAddRuleInput\",\"rule\",\"source\",\"vlan\"],\"extensions\":{\"code\":\"GRAPHQL_VALIDATION_FAILED\"}}],\"data\":null}"}
```

**Cause:**
When user specified `source = {}` (empty source), the hydration created a Source struct with all nil fields. Go's JSON marshaler serialized nil slices as `null`, which the API rejected.

**API Request (problematic):**
```json
{
  "source": {
    "floatingSubnet": null,
    "globalIpRange": null,
    "group": null,
    "host": null,
    "ip": null,
    "ipRange": null,
    "mac": null,
    "networkInterface": null,
    "site": null,
    "siteNetworkSubnet": null,
    "subnet": null,
    "systemGroup": null,
    "vlan": null
  }
}
```

**Root Cause Analysis:**
- Go struct fields with nil slices serialize to `null` in JSON
- The Cato API's GraphQL schema doesn't accept `null` for these array fields
- It expects either omitted fields OR empty arrays `[]`

**Solution:**
Initialize all list fields to empty slices in the hydration functions:

```go
// Initialize all list fields to empty slices to avoid null serialization
createInput.Vlan = make([]scalars.Vlan, 0)
updateInput.Vlan = make([]scalars.Vlan, 0)
createInput.Mac = make([]string, 0)
updateInput.Mac = make([]string, 0)
createInput.IP = make([]string, 0)
updateInput.IP = make([]string, 0)
// ... etc for all fields
```

**API Request (fixed):**
```json
{
  "source": {
    "floatingSubnet": [],
    "globalIpRange": [],
    "group": [],
    "host": [],
    "ip": [],
    "ipRange": [],
    "mac": [],
    "networkInterface": [],
    "site": [],
    "siteNetworkSubnet": [],
    "subnet": [],
    "systemGroup": [],
    "vlan": []
  }
}
```

**Files Modified:**
- `internal/provider/hydrate_socket_lan_firewall_rule_api.go` - Added initialization in `hydrateSocketLanFirewallSourceApi()` and `hydrateSocketLanFirewallDestinationApi()`

**Key Lesson:**
The Cato GraphQL API distinguishes between `null` and `[]` (empty array). Always initialize slice fields with `make([]Type, 0)` to ensure empty arrays are sent instead of null.

**Reference:**
The Socket LAN Network Rule (`hydrate_socket_lan_network_rule_api.go`) already had this pattern implemented correctly at lines 246-249.

---

## Error 4: GraphQL Validation - Application Cannot Be Null

**Date:** 2026-02-20

**Error Message:**
```
Error: Catov2 API PolicySocketLanFirewallAddRule error

{"networkErrors":{"code":422,"message":"Response body {\"errors\":[{\"message\":\"cannot be null\",\"path\":[\"variable\",\"socketLanFirewallAddRuleInput\",\"rule\",\"application\"],\"extensions\":{\"code\":\"GRAPHQL_VALIDATION_FAILED\"}}],\"data\":null}"}
```

**Cause:**
When user did not specify an `application` block, the field was left as `null`. The API requires application (and service) to be objects with empty arrays, not null.

**API Request (problematic):**
```json
{
  "rule": {
    "application": null,
    "service": null,
    "source": {...},
    "destination": {...}
  }
}
```

**Root Cause Analysis:**
- Unlike source/destination which were conditionally initialized, application and service were only initialized when specified by the user
- The API requires ALL of these objects to be present with empty arrays
- This is different from state hydration where null is acceptable

**Solution:**
Always initialize application and service objects with empty arrays, regardless of user input:

```go
// Application - always initialize to avoid null serialization
result.create.Rule.Application = &cato_models.SocketLanFirewallApplicationInput{
    Application:   make([]*cato_models.ApplicationRefInput, 0),
    CustomApp:     make([]*cato_models.CustomApplicationRefInput, 0),
    Domain:        make([]string, 0),
    Fqdn:          make([]string, 0),
    GlobalIPRange: make([]*cato_models.GlobalIPRangeRefInput, 0),
    IP:            make([]string, 0),
    IPRange:       make([]*cato_models.IPAddressRangeInput, 0),
    Subnet:        make([]string, 0),
}

// Service - always initialize to avoid null serialization
result.create.Rule.Service = &cato_models.SocketLanFirewallServiceTypeInput{
    Simple:   make([]*cato_models.SimpleServiceInput, 0),
    Standard: make([]*cato_models.ServiceRefInput, 0),
    Custom:   make([]*cato_models.CustomServiceInput, 0),
}
```

Also added helper functions for source/destination when not specified:
```go
// When source/destination not specified by user, still initialize
if !ruleData.Source.IsNull() {
    hydrateSocketLanFirewallSourceApi(...)
} else {
    initializeEmptySocketLanFirewallSource(createInput, updateInput)
}
```

**API Request (fixed):**
```json
{
  "rule": {
    "application": {
      "application": [],
      "customApp": [],
      "domain": [],
      "fqdn": [],
      "globalIpRange": [],
      "ip": [],
      "ipRange": [],
      "subnet": []
    },
    "service": {
      "simple": [],
      "standard": [],
      "custom": []
    },
    "source": {...},
    "destination": {...}
  }
}
```

**Files Modified:**
- `internal/provider/hydrate_socket_lan_firewall_rule_api.go`
  - Always initialize application and service with empty arrays
  - Added `initializeEmptySocketLanFirewallSource()` helper function
  - Added `initializeEmptySocketLanFirewallDestination()` helper function

**Key Lesson:**
The Socket LAN Firewall API requires ALL nested objects (source, destination, application, service) to be present with empty arrays. This is stricter than the Network Rule API which allows some fields to be null.

---

## Error 5: GraphQL Validation - Tracking Alert Fields Cannot Be Null

**Date:** 2026-02-20

**Error Message:**
```
Error: Catov2 API PolicySocketLanFirewallAddRule error

{"networkErrors":{"code":422,"message":"Response body {\"errors\":[{\"message\":\"cannot be null\",\"path\":[\"variable\",\"socketLanFirewallAddRuleInput\",\"rule\",\"tracking\",\"alert\",\"subscriptionGroup\"],\"extensions\":{\"code\":\"GRAPHQL_VALIDATION_FAILED\"}}],\"data\":null}"}
```

**Cause:**
When user specified `tracking.alert` with only `enabled` and `frequency`, the array fields (`subscriptionGroup`, `webhook`, `mailingList`) were left as null.

**API Request (problematic):**
```json
{
  "tracking": {
    "alert": {
      "enabled": false,
      "frequency": "DAILY",
      "mailingList": null,
      "subscriptionGroup": null,
      "webhook": null
    },
    "event": {"enabled": true}
  }
}
```

**Solution:**
Initialize alert array fields with empty slices when creating the alert object:

```go
createAlert := &cato_models.PolicyRuleTrackingAlertInput{
    Enabled:           alertData.Enabled.ValueBool(),
    SubscriptionGroup: make([]*cato_models.SubscriptionGroupRefInput, 0),
    Webhook:           make([]*cato_models.SubscriptionWebhookRefInput, 0),
    MailingList:       make([]*cato_models.SubscriptionMailingListRefInput, 0),
}
```

**API Request (fixed):**
```json
{
  "tracking": {
    "alert": {
      "enabled": false,
      "frequency": "DAILY",
      "mailingList": [],
      "subscriptionGroup": [],
      "webhook": []
    },
    "event": {"enabled": true}
  }
}
```

**Files Modified:**
- `internal/provider/hydrate_socket_lan_firewall_rule_api.go` - Initialize alert arrays with empty slices

**Key Lesson:**
Even deeply nested objects need their array fields initialized with empty slices. The pattern applies at all levels of the object hierarchy.

---

## Error 6: State Inconsistency - Service and Tracking Alert Objects

**Date:** 2026-02-20

**Error Messages:**
```
Error: Provider produced inconsistent result after apply
.rule.service: was null, but now cty.ObjectVal(map[string]cty.Value{...})

Error: Provider produced inconsistent result after apply
.rule.tracking.alert: was cty.ObjectVal(...), but now null.
```

**Cause:**
Two related state hydration issues:
1. **Service**: State hydration always created a service object even when empty, but plan expected null
2. **Tracking Alert**: State hydration only created alert when `Enabled = true`, but user specified alert with `enabled = false`

**Root Cause Analysis:**

**Service Issue:**
- User didn't specify `service` block (plan = null)
- State hydration created `service = { simple: null, standard: null, custom: null }`
- Result: Plan (null) != State (object)

**Alert Issue:**
- User specified `alert = { enabled = false, frequency = "DAILY" }`
- State hydration checked `if apiTracking.Alert.Enabled` which was false
- State hydration skipped creating the alert object, returned null
- Result: Plan (object) != State (null)

**Solution:**

**Service Fix** (in main hydration function):
```go
// Service - only create service object if there's actual content
if len(apiRule.Service.Simple) > 0 || len(apiRule.Service.Standard) > 0 || len(apiRule.Service.Custom) > 0 {
    result.Service = hydrateSocketLanFirewallServiceState(ctx, apiRule.Service)
} else {
    result.Service = types.ObjectNull(SocketLanFirewallServiceAttrTypes)
}
```

**Alert Fix** (always create event and alert objects):
```go
// Event - always create event object (API always returns tracking)
eventObj, _ := types.ObjectValue(TrackingEventAttrTypes, map[string]attr.Value{
    "enabled": types.BoolValue(apiTracking.Event.Enabled),
})
trackingAttrs["event"] = eventObj

// Alert - always create alert object with all fields (even when enabled = false)
alertAttrs := map[string]attr.Value{
    "enabled":            types.BoolValue(apiTracking.Alert.Enabled),
    "frequency":          types.StringNull(),
    "subscription_group": types.SetNull(NameIDObjectType),
    // ...
}
// ... populate fields ...
alertObj, _ := types.ObjectValue(TrackingAlertAttrTypes, alertAttrs)
trackingAttrs["alert"] = alertObj
```

**Files Modified:**
- `internal/provider/hydrate_socket_lan_firewall_rule_state.go`
  - Conditional service object creation
  - Always create event and alert objects

**Key Lessons:**
1. **For optional blocks with content** (like service): Only create in state when there's actual data
2. **For required/expected blocks** (like tracking.event, tracking.alert): Always create in state because user always specifies them, even with `enabled = false`
3. Don't use boolean values as conditions for creating objects - the presence of the config block determines whether to create the state object

---

## Summary of Patterns

### Pattern 1: Optional Nested Objects in State
Only create optional nested objects when they contain data:
```go
if hasContent {
    result.Field = hydrateFieldState(...)
} else {
    result.Field = types.ObjectNull(FieldAttrTypes)
}
```

### Pattern 2: Empty Slices vs Nil Slices for API
Always initialize slices to avoid null serialization:
```go
input.Field = make([]Type, 0)  // Serializes to []
// NOT: leave as nil           // Serializes to null
```

### Pattern 3: Schema Design for Optional Collections
Don't use `Computed: true` on optional collection fields unless the API computes values. Use validators to enforce constraints:
```go
"custom": schema.ListNestedAttribute{
    Optional: true,
    Validators: []validator.List{
        listvalidator.SizeAtLeast(1),  // Prevents custom = []
    },
}
```

---

## Related Files

- `internal/provider/resource_socket_lan_network_rule.go` - Network rule resource and schema
- `internal/provider/resource_socket_lan_firewall_rule.go` - Firewall rule resource and schema
- `internal/provider/hydrate_socket_lan_network_rule_api.go` - Network rule API hydration
- `internal/provider/hydrate_socket_lan_network_rule_state.go` - Network rule state hydration
- `internal/provider/hydrate_socket_lan_firewall_rule_api.go` - Firewall rule API hydration
- `internal/provider/hydrate_socket_lan_firewall_rule_state.go` - Firewall rule state hydration
- `internal/provider/hydrate_ifw_rule_state.go` - Reference implementation (Internet Firewall)
