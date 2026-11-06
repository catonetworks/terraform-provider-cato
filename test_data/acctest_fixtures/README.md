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
