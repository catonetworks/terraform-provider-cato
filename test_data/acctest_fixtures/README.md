# Acceptance account fixtures

Each CI account needs its own fixtures because any group can land on any account.
Use dedicated test accounts. Names beginning with `tf-fixture-` survive the test
cleanup, which removes disposable sites beginning with `acctest_`.

`main.tf` creates three sites (including native interfaces/ranges), hosts, DHCP
relay groups, mailing lists, subscription groups, and custom applications. `provision.py` also creates
three fixture users and disabled, unattached webhooks on `example.invalid`.
The users must be enabled and have remote-access eligibility to be returned by
the existing `vpnUser` lookup. It also seeds `acctest_advanced_group_000` through
`002`, the persistent names expected by the existing group fixture helper.
No passwords or API keys are issued for fixture users.
The account must have sufficient trial VPN-user capacity before provisioning.
Wi-Fi suites also require the `SOCKET_WIFI` migration. Enroll the dedicated
accounts through CMA's Tag Rollout with an explicit Include audience; this tag
has a migration action and cannot be added through Feature Tags.
Enable Static Range Translation under Resources > System Settings for the
socket-site translated-subnet test. IP Overlapping must remain disabled.

The user bootstrap calls the public API directly because a new account can
allocate user ID `0`, which the current Terraform user resource rejects.
Existing eligibility selections are retained. A populated account using ALL
allocation requires review rather than silently changing its eligibility.

Build the current provider and keep the binary, account file, logs, and state in
a private directory outside Git. For example, after exporting `PRIVATE_DIR` to
that directory and `CATO_BASEURL` to the environment's HTTPS GraphQL endpoint:

```sh
mkdir -p "$PRIVATE_DIR/provider"
go build -o "$PRIVATE_DIR/provider/terraform-provider-cato" .
python3 test_data/acctest_fixtures/provision.py \
  --accounts "$PRIVATE_DIR/accounts.json" \
  --state-dir "$PRIVATE_DIR/state" \
  --provider-dir "$PRIVATE_DIR/provider" \
  --endpoint "$CATO_BASEURL"
```

The account file uses the same array format as `CATO_ACCTEST_ACCOUNTS`. Provisioning
merges generated references into each record's `test_vars`. Credentials are
passed to Terraform through its environment, never arguments. The Terraform
configuration pins the published provider; the bootstrap deliberately uses the
built provider through a development override to include resources added since
the release. It does not run acceptance tests.

The following fixture types require CMA configuration. Create three of each per
account, then populate that account's `test_vars` with their actual IDs/names:

- `floating_ranges`: `tf-fixture-floating-0` through `2`, subnets `10.241.0.0/24`
  through `10.241.2.0/24`.
- `custom_categories`: `tf-fixture-category-0` through `2`, each referencing its
  own `tf-fixture-N.example.invalid` domain.
- `device_postures`: `tf-fixture-posture-0` through `2`, referencing an unattached
  Running Process check named `tf-fixture-check` for `tf-fixture.exe` on Windows.
  Keep account-wide device posture enforcement unchanged.
- `user_groups`: `tf-fixture-usergroup-0` through `2`; empty manual groups are
  sufficient references for the existing policy tests.
- `system_groups`: three existing built-in groups, such as All Sites, All SDP
  Users, and All Floating Ranges, verified in that account.

`custom_apps` comes from the Terraform output. Every reference is an object with
`id` and `name`. Do not copy IDs of account-owned fixtures from another account.
The initial ten CI accounts were provisioned through public APIs and CMA's
supported GraphQL console; their complete references are stored in the secret.

To update the secret from the protected local file without exposing keys in
arguments, shell history, or logs:

```sh
python3 -c 'import json, sys; print(json.dumps(json.load(sys.stdin), separators=(",", ":")))' \
  < "$PRIVATE_DIR/accounts.json" | \
  gh secret set CATO_ACCTEST_ACCOUNTS --repo catonetworks/terraform-provider-cato
```

Compact JSON avoids GitHub masking individual punctuation-only lines from a
pretty-printed JSON secret throughout the job logs. Individual API keys are
masked explicitly by the account-selection step.

## Optional live coverage

Each account record can contain a `test_env` object alongside `test_vars`.
The account selector exports only the fixture settings allowed in
`acctest_groups.py`; credentials, endpoints, runner paths, and multiline values
cannot be overridden through this object. Settings must reference the selected
account's fixtures.

To enable the additional suites on every randomly assigned account:

- Allocate a dedicated public IP within existing test-account capacity. IPsec,
  BGP, and remote-port-forward tests use the allocated-IP lookup.
- Enable `PPS_UA_TERMINAL_SERVER_UI` for Terminal Server. Enable
  `PPS_WEB_PROXY_UI` through a Tag Rollout with an explicit Include audience,
  because it has a migration action.
- Enable `ADD_SECONDARY_ON_HA_SITE_CREATION` and set
  `TFACC_ALLOW_UNREGISTERED_SECONDARY_SITE=1` for AWS/Azure secondary sites.
- Configure a dedicated static-IP subnet under IP Allocation Policy > Settings.
  Set `TFACC_SDP_STATIC_IP` and `TFACC_SDP_STATIC_IP_ANCHOR` to two distinct,
  unused addresses in it. The tests create disabled users and remove their rules.
- For DNS/DHCP, set `TFACC_ENABLE_DNS_DHCP_CRUD=true`,
  `TFACC_DNS_DHCP_ALLOWED_ACCOUNT_ID` to that record's ID, `CATO_TEST_SITE_ID` to
  its persistent fixture site, `TFACC_DNS_DHCP_SERVERS` to two test DNS addresses,
  and `TFACC_DHCP_OPTION_TAG` to a preferred unused option tag. If that tag is
  occupied, the suite selects a free private-use tag (224–254) and preserves
  existing options. The suite restores prior
  account settings and fixture-site collections. The explicit matching account
  permits dedicated HTTPS environments in addition to the original dev endpoint.
- For URL containers, set `TFACC_FQDN_CONTAINER_URL` and
  `TFACC_IP_RANGE_CONTAINER_URL` to public HTTPS feeds, and
  `TFACC_CONTAINER_NOTIFICATION_MAILING_LIST_ID` to a dedicated mailing list
  containing only a reserved `example.invalid` address. Set
  `TFACC_CONTAINER_URL_ALLOW_NOTIFICATIONS=1`. An empty list is rejected by the
  API. The test containers remain unattached to policies and are destroyed.

The dedicated pool was validated with a `/24` static subnet and two addresses
reserved for serial execution. DNS/DHCP, standalone SDP static-IP, and secondary
site coverage were exercised on each account; other newly enabled suites were
exercised on representative pool accounts.

Remaining prerequisites are explicit: account CRUD requires a reseller-capable
key; license/App Connector bandwidth tests require active pooled-bandwidth
licenses; physical Cloud Interconnect needs a PoP advertising the isolated
`CrossConnetMock` provider with `DOT1Q`. The trial pool has neither the licenses
nor an eligible mock PoP. Bulk Site Web Proxy rule creation remains blocked by
the backend rejecting `transportProtocol`, including requests that omit it;
policy access and standalone section CRUD are available. Keep that specific
bulk test excluded with its backend reason until the API supports creation.
