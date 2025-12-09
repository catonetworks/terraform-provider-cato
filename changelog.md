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

## 0.0.35 (2026-07-23)

### Features

- Updated bulk rule index resources for if_rule and wf_rule to map sections and rules by name in state, for deployment when source rules are purged for back/restore function

## 0.0.36 (2026-07-30)

### Features

- Updated network_range and socket_site resources to properly implement read operation to support import and write all available attributes from read back to state
- Updated all data sources using entityLookup to add 0 and limit to support large numbers of records in return
- Added network_range data source

## 0.0.37 (2026-07-31)

### Features

- Updated network_range to mark interface_id as required
- Updated ifw rule resource to set activeon to default values initially

## 0.0.39 (2026-08-01)

### Features

- updated socket_site resource to support additional outputs of default interface index and id, network_range to support interface_id, or interface_index
- Fixed license resource entityLookup to retrieve all sites to validate site ID, previously limited to 50 without limit of 0 in entityLookup

## 0.0.40 (2026-08-04)

### Features

- Minor update to license schema in SDK adding PENDING to enum for license status

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


## 0.0.42 (2025-09-03)

### Features
- Fixed index field in IFW and WAN rules to be computed 
- Updated socket_site to support native_network range name and default interface name, and validated import export for these fields

## 0.0.43 (2025-09-11)

### Features
- Fixed logic in network_range, socket_site resources to support proper import/export for all available attributes

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
 
## 0.0.46 (2025-09-24)
- Fix for destType mapping for cloud deployments

## 0.0.47 (2025-09-24)
- Minor fix for state drift for at.position for IF_RULES and WAN_RULES for imports vs create/update

## 0.0.48 (2025-10-28)
- Added support for isDefault for native range interface
- Fixed exceptions for IFW rules and WAN rules
- Fixed subnet predicate in source, destination and exceoption source/destination
- Fixed custom service port and port range issue 

## 0.0.49 (2025-10-29)
- Fixed drift issues for license resource attributes in state

## 0.0.50 (2025-10-30)
- Added updated version of the SDK include TLS Inpection and WAN Network rule operations
- Added TLS Inspection Rule and section resources  

## 0.0.51 (2025-13-03)
- Added support for WanNetwork rule resource, section and bulk move resource
- Fixed a few minor state management issues with IFW and WAN rules for device_attribues, custom_service, etc.

## 0.0.53 (2025-11-10)
- Update socket_site resource to support custom default lan interface index for the native range
- Added all enum values to dest type for socket site resource

## 0.0.54 (2025-11-13)
- Updated ipsec resource to support init_message, auth_message and network_ranges as well as added interface_id as native output for the resource. 

## 0.0.55 (2025-11-14)
- Added guide for managing site licenses in bulk from csv

## 0.0.56 (2025-12-08)
- Added resource for new advanced groups (cato_group, cato_group_members) along with data source 
- Added comprehensive unit tests for IFW and WAN FW resources, fixed exceptions  
- Updated readmes for index and for csv license import guide


