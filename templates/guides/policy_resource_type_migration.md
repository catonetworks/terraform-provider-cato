---
page_title: "Policy Resource Type Migration"
description: |-
  Migrate removed policy resource types to their full-name replacements.
---

# Policy Resource Type Migration

The 14 deprecated resource types listed below have been removed. These abbreviated types were available only in prerelease versions and were never included in a stable release. If you used them in a prerelease, update configurations and pinned modules to their full-name replacements before using this provider version. Bulk ordering state can migrate directly with a `moved` block; Private Access and Remote Port Forwarding require the import migration described below.

## Bulk ordering resources

- `cato_bulk_amfh_move_rule` → `cato_bulk_anti_malware_file_hash_move_rule`
- `cato_bulk_apc_move_rule` → `cato_bulk_app_control_move_rule`
- `cato_bulk_atr_move_rule` → `cato_bulk_app_tenant_restriction_move_rule`
- `cato_bulk_cc_move_rule` → `cato_bulk_client_connectivity_move_rule`
- `cato_bulk_dia_move_rule` → `cato_bulk_dynamic_ip_allocation_move_rule`
- `cato_bulk_rpf_move_rule` → `cato_bulk_remote_port_forward_move_rule`
- `cato_bulk_sb_move_rule` → `cato_bulk_socket_bypass_move_rule`
- `cato_bulk_ssia_move_rule` → `cato_bulk_sdp_static_ip_allocation_move_rule`
- `cato_bulk_st_move_rule` → `cato_bulk_split_tunnel_move_rule`
- `cato_bulk_swp_move_rule` → `cato_bulk_site_web_proxy_move_rule`
- `cato_bulk_ts_move_rule` → `cato_bulk_terminal_server_move_rule`

To adopt the full name with Terraform 1.8 or later, upgrade the provider, change the resource type and its references, and add a `moved` block in the module that declares the resource. Keep the ordering inputs the same during the move.

```terraform
moved {
  from = cato_bulk_cc_move_rule.policy_order
  to   = cato_bulk_client_connectivity_move_rule.policy_order
}
```

Run `terraform plan` and verify it reports the move and zero resources to add, change, or destroy. Apply that plan to record the new address in state. Changing the resource type without a move can cause Terraform to run the ordering and publication again. On Terraform versions earlier than 1.8, keep the provider pinned to a version that supports the abbreviated types until you can upgrade Terraform and perform the migration.

## Private Access and Remote Port Forwarding

- `cato_pa_section` → `cato_private_access_section`
- `cato_remote_port_fwd_rule` → `cato_remote_port_forward_rule`
- `cato_remote_port_fwd_section` → `cato_remote_port_forward_section`

The removed types had different schemas and policy publication behavior from their full-name replacements. These pairs require a configuration and import migration. For example, the full-name Remote Port Forwarding rule requires fields such as `description`, `remote_ips`, and `tracking` that could be omitted on the abbreviated resource.

Complete this migration while pinned to `= 2.0.0-prerelease.2`, which supports both names. With Terraform 1.7 or later, record the existing Cato rule or section ID and adapt the configuration to the full-name schema. Use a `removed` block with `destroy = false` for the old address and an `import` block with the existing ID for the new address. Confirm that the plan forgets the old address and imports the same Cato object. Review any proposed field or placement changes before applying. Once the state contains only the full-name resources, remove the completed `removed` blocks and upgrade the provider. See the [Terraform `removed` block reference](https://developer.hashicorp.com/terraform/language/block/removed) and the [Terraform `import` block reference](https://developer.hashicorp.com/terraform/language/block/import).
