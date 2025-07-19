# Changelog

## 0.0.19 (2025-04-23)

### Features
- Added full read support for IFW Rule and WANFW Rule in state
- Added resources for license, bgp_peer, and lan_interface
- Added data sources for license, and network_interface

## 0.0.21 (2025-04-23)

### Features
- Fixed documentation to include comprehensive examples for WAN FW rule
- Updated license resource fixing pooled bw logic for existing licensed sites
- Updated license data source to retrieve all CATO_SITE licenses with filters properly

## 0.0.22 (2025-04-28)

### Features
- Updated license resource to allow license assignment for same license on same site.  Also fixed state change issue for license_info attrobites while migrating to new license ID.
- Fixed WF and IF rule section resources to check for existing sections and throw errors properly
- Added city to socket_site resource now the that API supports it to update city in site_location, added debug as well to lan_interface

## 0.0.23 (2025-05-05)

### Features
- Updated siteLocation schema definition to optimize siteLocation validation for upcoming sitelocaions bulk csv module
- Updated all resources to include consistent debug for request and response of API calls to SDK
- Added city attribute to IPSec site resource

## 0.0.24 (2025-05-19)

### Features

- Updated data source for networkInterfaces to support querying for LAN on X1600, X1600_LTE, and X1700, as well as LAN1 on X1500 to fix socket module behavior 
- Updated socket lan interface resource to fail gracefully when trying to delete native range interface
- Updated docs for all data source to include examples and descriptions

## 0.0.27 (2026-06-23)

### Features

- Added data source for wan and internet firewalls
- Optimized the sitelocation data source to resolve immediately from index for exact matche searches
- Added internetOnly boolean for network_ranges

## 0.0.28 (2026-07-04)

### Features

- Added dhcpRelayGroup data source
- Updated if_rule and wf_rule to fix issues with tracking changes in exceptions and index field
- Fixed state in section resources to write id and attributes correctly to state

## 0.0.28 (2026-07-05)

### Features

- Fixed issues with resource_wan_fw_rules_index to address null pointers

## 0.0.32 (2026-07-16)

### Features

- Fixed issues with resource_wan_fw_rules_index to address null pointers
- Added data sources for ifRuleSections and wfRuleSections, fixed encoding issues in siteLocation and optimized internet_fw_rules_index and wan_fw_rules_index
- Fixed logic in move rule index resources to fail of invalid section id specified for section_to_start_after_id
- Updated socket_site, network_range, and network_interface resources to fix read operations and support import
- Fixed wan_interface resource to support import and read operation

## 0.0.33 (2026-07-18)

### Features

- Fixed issues in wan_interface
- Fixed issues with socket_site address and city fields with state
- Added resources for account and admin
