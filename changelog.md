# Changelog

## 0.0.90 (2026-07-08)

### Fixed
- Fixed false-positive `cato_socket_site` DHCP plan modifier errors by allowing `dhcp_microsegmentation=false` for non-`DHCP_RANGE` types and by only treating relay group fields as user-set when they differ from prior state.
- Fixed false-positive `interface_id`/`interface_index` "cannot be set simultaneously" errors in `cato_network_range` caused by Terraform propagating prior-state values for Optional+Computed attributes.
- Stabilized WAN interface hydration.

### Changed
- Added account snapshot caching in the provider to reduce redundant API calls.

### Tests
- Added unit test coverage for the DHCP relay and network-range explicit-field helpers, the account snapshot cache, and WAN interface hydration.

## 0.0.89 (2026-07-01)

### Fixed
- Fixed old state handling in the DHCP settings plan modifier and network range resource to correctly process prior state during plan evaluation.

### Changed
- Updated retry policy to use the GraphQL-aware `cato.BaseRetryPolicy` in the provider configuration.

### Tests
- Added unit test coverage for the DHCP settings plan modifier and `cato_network_range` resource.
- Updated private access policy acceptance tests for origins handling.
- Increased timeout in the flaky acceptance test runner script.

## 0.0.88 (2026-06-25)

### Added
- Added App Control resources: `cato_app_tenant_restriction_rule`, `cato_app_tenant_restriction_section`, `cato_application_control_policy`, `cato_application_control_rule`, and `cato_application_control_section`, including schema, hydration helpers, provider registration, examples, documentation, and acceptance tests.

### Fixed
- Fixed `cato_socket_site` DHCP settings null inconsistency after apply via a new plan modifier.

## 0.0.87 (2026-06-23)

### Added
- Added `cato_global_ip_ranges` resource support, including schema, validators, provider registration, examples, documentation, and acceptance tests.
- Added regression test support, regression environment files, network suite coverage updates, and acctest mock skill guidance.

### Changed
- Updated internet and WAN firewall rules index handling to use the `reorderPolicy` API for bulk rule reordering, with policy revision retry and cleanup support.
- Updated generated provider documentation.
- Updated the Cato Go SDK dependency.

### Fixed
- Fixed `translated_subnet` API payload handling for network range and LAN interface resources so it is omitted unless configured.
- Fixed ID/name handling in `cato_if_rule`.

## 0.0.86 (2026-06-17)

### Fixed
- Fixed `cato_socket_site` updates so `translated_subnet` is omitted from `siteUpdateNetworkRange` and `siteUpdateSocketInterface` when it is not set in Terraform config, even if plan/state still carry an API-hydrated value (accounts with Static Range Translation disabled).

## 0.0.85 (2026-06-16)

### Added
- Added acceptance test coverage for account, BGP peer, internet firewall rule indexes, TLS rule indexes, WAN firewall rule indexes, and WAN network rule indexes.

### Changed
- Refactored `cato_network_range` with shared DHCP settings helpers, extracted validators, and streamlined CRUD flow.
- Updated the Cato Go SDK dependency.

### Fixed
- Fixed `translated_subnet` handling for socket site and network range API payloads to omit empty or unset values.

## 0.0.84 (2026-06-05)

### Fixed
- Fixed bulk firewall rule reorder computed-state handling.

## 0.0.82 (2026-06-03)

### Changed
- Refactored socket site handling and related validators for native ranges, DHCP settings, interface indexes, destination types, and site connection types.

### Fixed
- Fixed HA support for AWS, GCP, and Azure socket sites.
- Fixed native range validation when Terraform values are unknown.
- Fixed socket site location hydration when `cityName` is null.

## 0.0.81 (2026-05-29)

### Fixed
- Fixed internet firewall rule state drift for service, users_group, and bulk metadata handling.
- Hardened internet firewall object-reference validation and improved service state compatibility.

## 0.0.80 (2026-05-28)

### Fixed
- Fixed internet firewall rule app reference handling to fail fast on invalid object references instead of emitting partial payloads.
- Fixed internet firewall rule refresh/apply consistency by preserving `rule.service = null` when no service values are returned.
- Added regression unit and acceptance test coverage for IFW invalid destination app references and empty-service refresh behavior.

## 0.0.79 (2026-05-26)

### Fixed
- Fixed city and WAN precedence regressions.
- Fixed DHCP relay group handling.
- Fixed linter issues.

### Changed
- Improved the unit test skill guidance.

## 0.0.78 (2026-05-20)
- Fixed network range drift and added unit test coverage for the updated behavior.
- Fixed internet firewall rule device attribute handling to ensure correct state and API mapping.


## 0.0.77 (2026-05-20)
- Fixed IF rule device attributes.


## 0.0.76 (2026-05-19)
- Fixed ID/name handling for internet firewall rule exceptions and added focused parser unit coverage.
- Simplified internet firewall rule reference parsing by moving shared logic into the parser package.
- Added release preparation and publishing skills for the provider release workflow.
- Added unit test authoring guidance and reference material for provider resource and data source tests.
- Added and refined agent workflow documentation, including testing, vulnerability checks, skill paths, and acceptance-test guidance.

## 0.0.75 (2026-05-18)
- Fixed license handling for accounts with more than 1,000 sites and added defensive unit coverage.
- Fixed `translated_subnet` handling for network range, LAN interface, and socket site native ranges to submit nil values correctly when unset.
- Added broader Terraform acceptance test coverage across provider resources and cleanup workflows.
- Updated the Cato Go SDK dependency.
- Hardened socket site update flows with bounded retries for transient backend conflicts and improved connection type hydration.
- Updated internet and WAN firewall rule hydration to send empty API objects/lists instead of null values for service and action configuration fields.

## 0.0.74 (2026-05-11)
- Fixed `translated_subnet` handling for socket site native ranges to avoid state drift and preserve API-hydrated values correctly.

## 0.0.73 (2026-05-05)
- Removed length validator from Site Location Fields.

## 0.0.72 (2026-04-30)
- Removed rules index from terraform state.

## 0.0.71 (2026-04-22)
- Added configurable provider retry settings
- Improved the siteLocation data source read path and added tests to better validate API-backed state handling
- Fixed internet firewall rule index known-after-apply behavior

## 0.0.70 (2026-04-14)
- Refactored network_range DHCP settings hydration to properly support import by handling null/unknown state, and always setting DhcpMicrosegmentation from API response for all DHCP types
- Network_range now sets dhcp_settings to null for Routed and Direct range types (API returns values but they are not valid to submit)
- Network_range DHCP_RELAY now writes both relay_group_id and relay_group_name to state to prevent drift regardless of which field the user specifies in config
- Added getSiteIdFromNetworkRange helper using entityLookup to resolve site_id and interface_id during network_range import when not present in state
- Fixed socket_site getNativeRange to filter by matching subnet, preventing overwrite from other native-type ranges (e.g., LAN2 with is_native_range=TRUE)
- Simplified socket_site translated_subnet handling to always hydrate directly from API, removing complex plan/state comparison logic
- Simplified socket_site site location hydration to always use API values for state_code, address, and city during import and refresh
- Socket_site microsegmentation now always set from API response for proper hydration during import

## 0.0.69 (2026-04-10)
- Added in new getSiterangeList API moving off of entityLookup for network_range and socket_site resources. 
- Updated socket_site resource to accommodate state chagnes with default values for DHCP for native range

## 0.0.66 (2026-03-30)
- Adding SDK trace_id error logging with request body, rereleasing 0.0.66

## 0.0.65 (2026-03-20)
- FIxed issues in lan firewall, siteLocation and socket site logic (swap interface call)

## 0.0.64 (2026-02-24)
- Added resources for socketLan network and firewall with section.
- Updated network_range to support microsegmentation for DHCP_RELAY (previously not supported by API)
- Updated socket_site resource to read timezone from API instead of from local dataset, and fixed logic around interface mapping
- Reverted siteLocation data to use local json dataset
- Added fix for DHCP_SETTINGS as null in socket_site

## 0.0.58 (2025-12-10)
- Updated socket_site, lan_interface, wan_interface, and network_range resources to add all read attributes now available in the API to hystrate state.
- Added data source for host lookup

## 0.0.57 (2025-12-10)
- Updated group data-source to support id, and name filters, and updated readmes for new resources to provide examples for all

## 0.0.56 (2025-12-08)
- Added resource for new advanced groups (cato_group, cato_group_members) along with data source 
- Added comprehensive unit tests for IFW and WAN FW resources, fixed exceptions  
- Updated readmes for index and for csv license import guide

## 0.0.55 (2025-11-14)
- Added guide for managing site licenses in bulk from csv

## 0.0.54 (2025-11-13)
- Updated ipsec resource to support init_message, auth_message and network_ranges as well as added interface_id as native output for the resource. 

## 0.0.53 (2025-11-10)
- Update socket_site resource to support custom default lan interface index for the native range
- Added all enum values to dest type for socket site resource

## 0.0.51 (2025-13-03)
- Added support for WanNetwork rule resource, section and bulk move resource
- Fixed a few minor state management issues with IFW and WAN rules for device_attribues, custom_service, etc.

## 0.0.50 (2025-10-30)
- Added updated version of the SDK include TLS Inpection and WAN Network rule operations
- Added TLS Inspection Rule and section resources  

## 0.0.49 (2025-10-29)
- Fixed drift issues for license resource attributes in state

## 0.0.48 (2025-10-28)
- Added support for isDefault for native range interface
- Fixed exceptions for IFW rules and WAN rules
- Fixed subnet predicate in source, destination and exceoption source/destination
- Fixed custom service port and port range issue 

## 0.0.47 (2025-09-24)
- Minor fix for state drift for at.position for IF_RULES and WAN_RULES for imports vs create/update

## 0.0.46 (2025-09-24)
- Fix for destType mapping for cloud deployments

## 0.0.45 (2025-09-24)

### Features
- **NEW RESOURCE**: Added `cato_lan_interface_lag_member` resource for managing LAN LAG member interfaces on socket sites
- **SOCKET_SITE ENHANCEMENTS**: Added LAG (Link Aggregation Group) support to `cato_socket_site` resource:
  - Added `lag_min_links` attribute to specify minimum number of interfaces for LAG configuration
  - Added `interface_dest_type` attribute with support for LAN, LAN_LAG_MASTER, and LAN_LAG_MASTER_AND_VRRP
  - Added validation to ensure LAG configuration consistency between `interface_dest_type` and `lag_min_links`
- **LAN_INTERFACE ENHANCEMENTS**: Enhanced `cato_lan_interface` resource with LAG support:
  - Added `lag_min_links` attribute for LAG configuration
  - Improved state management for LAG-enabled interfaces
- **BUG FIXES**: 
  - Fixed "Provider produced inconsistent result after apply" error in socket_site resource by ensuring consistent use of plan values instead of state values during updates
  - Improved socket interface update logic to prevent subnet value mismatches
  - Enhanced debug logging for socket interface operations
- **DOCUMENTATION**: Updated resource documentation for socket_site and lan_interface to include new LAG attributes and examples
 
## 0.0.43 (2025-09-11)

### Features
- Fixed logic in network_range, socket_site resources to support proper import/export for all available attributes

## 0.0.42 (2025-09-03)

### Features
- Fixed index field in IFW and WAN rules to be computed 
- Updated socket_site to support native_network range name and default interface name, and validated import export for these fields

## 0.0.41 (2025-08-19)

### Features

- Updated IFW and WAN firewall rules to support device_attributes with category, type, model, manufacturer, os, and osVersion filtering
- Added active_period support for IFW and WAN rules with effective_from and expires_at functionality 
- Enhanced exceptions handling for firewall rules with comprehensive device attribute filtering
- Added new plan modifiers for better state management:
  - active_period_modifier for handling rule activation periods
  - ifw_exceptions_set_modifier and wan_exceptions_set_modifier for exception handling
  - source_dest_object_modifier for source/destination object state management
  - empty_set_default_modifier for default empty set handling
- Added utility functions for IFW and WAN rule management in resource_internet_fw_utils.go and resource_wan_utils.go
- Enhanced socket_site resource with improved state management and additional interface outputs
- Updated license resource to fail gracefully when trial license is used on sites
- Added semantic equality checking for WAN firewall rules to prevent unnecessary updates
- Improved rule state hydration with better device attribute and exception handling
- Enhanced documentation and examples for IFW and WAN rule resources
- Updated bulk sites functionality for improved site management

## 0.0.40 (2025-08-04)

### Features

- Minor update to license schema in SDK adding PENDING to enum for license status

## 0.0.39 (2025-08-01)

### Features

- updated socket_site resource to support additional outputs of default interface index and id, network_range to support interface_id, or interface_index
- Fixed license resource entityLookup to retrieve all sites to validate site ID, previously limited to 50 without limit of 0 in entityLookup

## 0.0.37 (2025-07-31)

### Features

- Updated network_range to mark interface_id as required
- Updated ifw rule resource to set activeon to default values initially

## 0.0.36 (2025-07-30)

### Features

- Updated network_range and socket_site resources to properly implement read operation to support import and write all available attributes from read back to state
- Updated all data sources using entityLookup to add 0 and limit to support large numbers of records in return
- Added network_range data source

## 0.0.35 (2025-07-23)

### Features

- Updated bulk rule index resources for if_rule and wf_rule to map sections and rules by name in state, for deployment when source rules are purged for back/restore function

## 0.0.33 (2025-07-18)

### Features

- Fixed issues in wan_interface
- Fixed issues with socket_site address and city fields with state
- Added resources for account and admin

## 0.0.32 (2025-07-16)

### Features

- Fixed issues with resource_wan_fw_rules_index to address null pointers
- Added data sources for ifRuleSections and wfRuleSections, fixed encoding issues in siteLocation and optimized internet_fw_rules_index and wan_fw_rules_index
- Fixed logic in move rule index resources to fail of invalid section id specified for section_to_start_after_id
- Updated socket_site, network_range, and network_interface resources to fix read operations and support import
- Fixed wan_interface resource to support import and read operation

## 0.0.28 (2025-07-05)

### Features

- Fixed issues with resource_wan_fw_rules_index to address null pointers

## 0.0.28 (2025-07-04)

### Features

- Added dhcpRelayGroup data source
- Updated if_rule and wf_rule to fix issues with tracking changes in exceptions and index field
- Fixed state in section resources to write id and attributes correctly to state

## 0.0.27 (2025-06-23)

### Features

- Added data source for wan and internet firewalls
- Optimized the sitelocation data source to resolve immediately from index for exact matche searches
- Added internetOnly boolean for network_ranges

## 0.0.24 (2025-05-19)

### Features

- Updated data source for networkInterfaces to support querying for LAN on X1600, X1600_LTE, and X1700, as well as LAN1 on X1500 to fix socket module behavior 
- Updated socket lan interface resource to fail gracefully when trying to delete native range interface
- Updated docs for all data source to include examples and descriptions

## 0.0.23 (2025-05-05)

### Features
- Updated siteLocation schema definition to optimize siteLocation validation for upcoming sitelocaions bulk csv module
- Updated all resources to include consistent debug for request and response of API calls to SDK
- Added city attribute to IPSec site resource

## 0.0.22 (2025-04-28)

### Features
- Updated license resource to allow license assignment for same license on same site.  Also fixed state change issue for license_info attrobites while migrating to new license ID.
- Fixed WF and IF rule section resources to check for existing sections and throw errors properly
- Added city to socket_site resource now the that API supports it to update city in site_location, added debug as well to lan_interface

## 0.0.21 (2025-04-23)

### Features
- Fixed documentation to include comprehensive examples for WAN FW rule
- Updated license resource fixing pooled bw logic for existing licensed sites
- Updated license data source to retrieve all CATO_SITE licenses with filters properly

## 0.0.19 (2025-04-23)

### Features
- Added full read support for IFW Rule and WANFW Rule in state
- Added resources for license, bgp_peer, and lan_interface
- Added data sources for license, and network_interface
