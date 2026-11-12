# cato_webhook

Status: implemented 2026-11-11; unit-tested and `go build`/`go vet` clean. Live acceptance
run attempted against an authorized dev tenant (see "Live acceptance" below for the exact
result) -- do not claim passing acceptance coverage beyond what was actually observed there.

- Package: `internal/provider/account/webhook/`.
- Template: `internal/provider/account/mailinglist/` (file layout, write-pattern, error
  sniffing); `internal/provider/network/secondaryawsvsocket/` (single nested id/name-style
  ref attribute pattern, though this resource ended up using a simpler `{by,input}` shape --
  see "related_webhook_ref" below).
- Creation input: `CreateWebhookInput`. Update input: `UpdateWebhookInput` (carries a
  `webhook: WebhookIntegrationRefInput!` target selector -- the ID being updated). Delete
  input: `WebhookIntegrationRefInput`. Import target: `id`.

## API operations

- SDK (cato-go-sdk) mutation methods exist as of this change (added via `sources/*.gql` +
  `make`/`go tool gqlgenc`, the generated/official path -- see cato-go-sdk's own
  `operations/manifest.json` for provenance):
  - `Catov2.NotificationIntegrationCreateWebhook`
  - `Catov2.NotificationIntegrationUpdateWebhook`
  - `Catov2.NotificationIntegrationDeleteWebhook`
  - `Catov2.NotificationIntegrationTestWebhook` (generated for completeness; **not** used by
    this resource's CRUD -- `testWebhook` is not a Terraform lifecycle operation).
  - The 3 webhook QUERY methods (`NotificationIntegrationWebhook`,
    `NotificationIntegrationWebhookList`, `NotificationIntegrationWebhookFields`) already
    existed in the SDK before this change.
- **This resource's `client_webhook.go` does NOT call any of the generated Catov2 methods
  above.** It hand-writes its own minimal inline GraphQL documents (mirroring the
  mailing_list/subscription_group convention: SDK access isolated in `client_*.go`, inline
  doc constants, `c.client.V2().Client.Post(...)`) for all four operations plus the
  single-webhook read. Two independent reasons converge on this:
  1. **Established repo convention.** `mailing_list`'s own generated mutation methods exist
     in the SDK too, and its resource still reimplements them inline with a narrower field
     selection than the generated/CLI document ("Request only fields owned by this
     resource" -- see `mailinglist/client_mailing_list.go`). This resource follows the same
     pattern: `createWebhook`/`updateWebhook` select only `webhook { id name }`,
     `deleteWebhook` selects only `webhook { id }`.
  2. **A live-API-breaking field in the generated/CLI selection.** The CLI-sourced
     `mutation.notification.integration.{create,update,delete}Webhook.txt` payloads (and
     therefore what a literal `operations-import` would have produced, and what the
     pre-existing `query.notification.integration.webhook{,List}.gql` SDK sources already
     select) all include `auth { ... on CustomHeadersAuth { plainHeaders{...}
     secretHeaders{...} } }`. `CustomHeadersAuth` is a real type in the bundled schema
     (`cato_api.graphqls`, `AuthInterface` implementors: BasicAuth, BearerTokenAuth,
     CustomHeaderAuth, CustomHeadersAuth, NoAuth) but is **not yet deployed** on every
     tenant's `notification` microservice; selecting it fails at runtime with
     `Validation error (UnknownType@[...auth]) : Unknown type 'CustomHeadersAuth'`
     (`DOWNSTREAM_SERVICE_ERROR`, serviceName `notification`). This was reported mid-task by
     a peer session doing a live diagnosis against the same API and independently confirmed
     here by inspecting the actual CLI payload files
     (`cato-cli/queryPayloads/mutation.notification.integration.*Webhook.txt`), which all
     carry that exact fragment.
  - **Consequence:** this resource's Create/Update/Delete documents, and its single-webhook
    Read document, never select `auth` at all (not even the safe `auth { authType }` scalar
    -- since `auth` is write-only and never parsed from any API response here anyway, there
    is no reason to request it). The **data source** (`cato_webhooks`, a different package)
    applies the identical fix for its list query. This is a deliberate deviation from the
    blueprint's stated default ("data sources call generated Catov2 methods directly") --
    documented there too (`internal/provider/account/data_sources/webhook/` has no
    `ai_context.md` of its own; see the comment above `webhookListDTO` in
    `datasource_webhooks.go`).
  - The 4 generated Catov2 mutation methods still exist in the SDK (satisfying the
    "generate the mutations the official way" requirement and giving the SDK first-class,
    reusable operations for any future caller), they are simply unused by this resource.

## Write-only fields: body, auth, related_webhook_ref

- `body` (Sensitive) and `auth` (all three sub-variants' secret fields Sensitive) are never
  refreshed from the API on Read. The Cato notification API does not echo back webhook
  secrets, and `body` is treated the same way defensively (a payload template can embed
  secrets inline, and the `catocli export webhooks` convention already deliberately omits
  `body`/`auth` from exports for the same reason).
- **Mechanism:** rather than a plan-modifier/"use state if API returns empty" dance, `Read`
  simply never attempts to populate `body`/`auth`/`related_webhook_ref` from the API
  response at all -- `parseWebhook` (`converters_webhook.go`) copies these three straight
  from `prior` (the existing state) and only refreshes the remaining fields
  (`name`/`url`/`enabled`/`request_method`/`header`/`response_mapping`) from the API. Since
  all three are schema-Required-or-Optional and **not Computed**, Terraform's planned value
  for them is always taken from config regardless of what Read produces -- so this achieves
  the desired "write-only, no perpetual drift, but still reacts to a real config change"
  behavior with no custom plan modifier needed. Every other (non-secret) field uses the
  ordinary Read-refreshes-state behavior.
- `related_webhook_ref` gets the same treatment for a different, non-secret reason: the API
  only returns a plain `relatedWebhookId` scalar on read (`Webhook.relatedWebhookId: ID`),
  never the referenced webhook's name, so a name-only reference cannot be round-tripped from
  a Read response. Modeling it as Optional-only (not Computed) and carrying it forward sidesteps the problem entirely.

## related_webhook_ref shape (deviation from the blueprint's idname suggestion)

The blueprint offered two options: mirror `shared/idname` (`{id, name}`, used by
mailing_list's `admin` and subscription_group's `integration`/`mailing_list`) or model a
simple `{by, input}` object matching `WebhookIntegrationRefInput` 1:1. This resource uses the
**latter**: `related_webhook_ref = { by = "ID"|"NAME", input = "<id-or-name>" }`, built via
a tiny hand-rolled `buildRelatedWebhookRef` in `converters_webhook.go` (no `idname` helper
call). Reasons:
1. `WebhookIntegrationRefInput{By, Input}` already matches the generic `idRefInputs`
   constraint in `shared/idname/convert_id_name.go`, so `idname.PrepareIDName` *could* build
   it -- but `idname.ParseIDName`/`ParseIDNameSet` (the Read-side half of that package)
   require an API type shaped like `{ID, Name}`, which `Webhook.relatedWebhookId` (a bare
   `*string`) is not. Using the package for only the write half and hand-rolling the read
   half (which is carried-forward anyway, see above) would be more code for no benefit.
2. The companion reference-architecture repo's `1-dependencies/bulk_webhooks.tf` plan
   (out of scope for this change, Part C of the parent plan) already expects exactly this
   `{by, input}` shape (`related_webhook_ref = { by = "NAME", input = ... }`), so matching it
   here avoids a mismatch for whoever implements that wiring next.

## Live-verified: the CustomHeadersAuth bug is real, today, on badev4

Independent confirmation beyond the peer-session report and the CLI payload inspection
above: running `catocli query notification integration webhookList '{}'` directly against
badev4 (profile `badev4`, account 24425) returns a hard `DOWNSTREAM_SERVICE_ERROR`:
`"Validation error (UnknownType@[notification/integration/webhookList/items/auth]) :
Unknown type 'CustomHeadersAuth'"`. This is catocli's own (CLI-sourced, full-selection)
query failing on a live, real, currently-authorized tenant -- not a hypothetical. It
confirms both that the bug is real right now and that this resource/data source's decision
to never select `auth` at all is the correct, necessary fix, not over-caution.

## Live acceptance

**PASSED** against badev4 (profile `badev4`, account 24425) on 2026-11-11:
`CATO_PROFILE=badev4 TF_ACC=1 DISABLE_POLICY_RULE_CLEANUP=true go test -tags acctest
-count=1 -timeout=5m -run TestAccWebhook -v ./internal/acctests/webhook/...` -> `--- PASS:
TestAccWebhook (8.49s)`. Exercised: create (bearer auth) -> plan shows no diff -> import ->
update (name + url) -> plan shows no diff -> destroy -> `CheckDestroy` confirms the webhook
is gone (a live "not found" GraphQL error for the deleted ID). No webhook was left behind on
badev4.

The first run of this exact test (same build, before the fix below) failed only in
`CheckDestroy`'s own error-message pattern match -- the resource's CRUD lifecycle itself
already worked correctly on that run too. The real not-found message, read directly off the
wire, is:

```
GQL-User-Error: Account ID [24425] subscription [16698] not found
```

Note **"subscription"**, not "webhook" -- the notification-integration backend uses the
same generic wording for every integration type (mailing lists, subscription groups,
webhooks). `missingWebhook` (`client_webhook.go`) and the acceptance test's `missingFixture`
were both updated to this verified string (originally guessed as `"webhook [%s] not
found"`, which was wrong). This is now a live-verified format, not a guess -- see
`missingWebhook`'s doc comment.
