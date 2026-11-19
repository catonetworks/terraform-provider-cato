---
page_title: "Policy Resource Type Migration"
description: |-
  Move bulk ordering resources to full names and identify replacements for deprecated policy types.
---

# Policy Resource Type Migration

Existing configurations and pinned modules can continue to use the abbreviated resource types. New configurations can use the full names.

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

Run `terraform plan` and verify it reports the move and zero resources to add, change, or destroy. Apply that plan to record the new address in state. Changing the resource type without a move can cause Terraform to run the ordering and publication again. On Terraform versions earlier than 1.8, continue using the abbreviated type until you can upgrade.

## Private Access and Remote Port Forwarding

- `cato_pa_section` → `cato_private_access_section`
- `cato_remote_port_fwd_rule` → `cato_remote_port_forward_rule`
- `cato_remote_port_fwd_section` → `cato_remote_port_forward_section`

The deprecated types remain available for existing configurations. The full-name types have their own schemas and policy publication behavior. These pairs require a configuration and import migration. For example, the full-name Remote Port Forwarding rule requires fields such as `description`, `remote_ips`, and `tracking` that can be omitted on the abbreviated resource.

To migrate one of these resources with Terraform 1.7 or later, record the existing Cato rule or section ID and adapt the configuration to the full-name schema. Use a `removed` block with `destroy = false` for the old address and an `import` block with the existing ID for the new address. Confirm that the plan forgets the old address and imports the same Cato object. Review any proposed field or placement changes before applying. See the [Terraform `removed` block reference](https://developer.hashicorp.com/terraform/language/block/removed) and the [Terraform `import` block reference](https://developer.hashicorp.com/terraform/language/block/import).
