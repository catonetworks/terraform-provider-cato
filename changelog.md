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