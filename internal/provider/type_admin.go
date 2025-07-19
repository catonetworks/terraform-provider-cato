package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type Admin struct {
	Id                   types.String `tfsdk:"id"`
	AdminId              types.String `tfsdk:"admin_id"`
	AccountId            types.String `tfsdk:"account_id"`
	Email                types.String `tfsdk:"email"`
	FirstName            types.String `tfsdk:"first_name"`
	LastName             types.String `tfsdk:"last_name"`
	PasswordNeverExpires types.Bool   `tfsdk:"password_never_expires"`
	MFAEnabled           types.Bool   `tfsdk:"mfa_enabled"`
	ManagedRoles         types.Set    `tfsdk:"managed_roles" json:"managed_roles,omitempty"`
	ResellerRoles        types.Set    `tfsdk:"reseller_roles" json:"reseller_roles,omitempty"`
}

type ManagedRole struct {
	ID   types.String `tfsdk:"id" json:"id,omitempty"`
	Name types.String `tfsdk:"name" json:"name,omitempty"`
}

type ResellerRole struct {
	ID              types.String `tfsdk:"id" json:"id,omitempty"`
	Name            types.String `tfsdk:"name" json:"name,omitempty"`
	AllowedAccounts types.Set    `tfsdk:"allowed_accounts" json:"allowed_accounts,omitempty"`
}
