# cato_dhcp_settings

Resource package: `dhcpsettings`. API-shaped schema approved by contributor.

- Read: `networkConfig.dhcp.settings`
- Create/upsert: `networkConfig.dhcp.updateSettings`
- Update: `networkConfig.dhcp.updateSettings`
- Delete: No remote delete: remove Terraform ownership and retain settings.
- Import: `ACCOUNT_ID`; validate account ownership as appropriate.
- Standalone APIs only; no `accountUpdate` or profiles. Collection resources manage one item; client adapters wrap API bulk envelopes.
- Suffixes, forwarding servers and relay servers are ordered lists. CC2 uses the first relay server for older sockets.
- Optional strings/bools/collections normalize to empty values in Terraform. DNS empty servers map to explicit API null; suffix uses a non-null empty list.
- DHCP lease/relay/group/timeout omission preserves the existing configuration. The API cannot clear selected group/timeout; disable relay instead. These Optional+Computed fields intentionally differ from generic empty defaults.
- Site and relay-group references use shared ID/name schema, plan modifiers and converters. A site change requires replacement.
- SDK interfaces follow repository AGENTS.md ResourceClient/resourceAPI conventions; response presence checks stay in DNS/DHCP adapters.
- Schema baseline fetched from `https://system.cc.catonetworks.com/api/schema?with_undocumented=true`.

The guide names NullBoolModifier/NullListModifier, which are not implemented in this checkout. Standard framework false/empty static defaults provide the approved equivalent behavior without adding shared infrastructure.
