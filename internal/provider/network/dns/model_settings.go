package dns

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type SettingsModel struct {
	ID                                types.String `tfsdk:"id"`
	PrimaryServer                     types.String `tfsdk:"primary_server"`
	SecondaryServer                   types.String `tfsdk:"secondary_server"`
	Suffix                            types.List   `tfsdk:"suffix"`
	AcceptDNSRequestsOnLanInterfaceIP types.Bool   `tfsdk:"accept_dns_requests_on_lan_interface_ip"`
}

type SiteSettingsModel struct {
	ID              types.String `tfsdk:"id"`
	SiteID          types.String `tfsdk:"site_id"`
	PrimaryServer   types.String `tfsdk:"primary_server"`
	SecondaryServer types.String `tfsdk:"secondary_server"`
	Suffix          types.List   `tfsdk:"suffix"`
}

func dnsAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": nc.IDAttribute(),
		"primary_server": schema.StringAttribute{Optional: true, Description: "Primary DNS server. Omission clears it; secondary server and " +
			"suffixes require a primary.", Validators: []validator.String{nc.IPValidator{}}},
		"secondary_server": schema.StringAttribute{Optional: true,
			Description: "Secondary DNS server. Omission clears it.",
			Validators:  []validator.String{nc.IPValidator{}}},
		"suffix": schema.ListAttribute{Optional: true,
			Computed:    true,
			ElementType: types.StringType,
			Default: listdefault.StaticValue(types.ListValueMust(types.StringType,
				nil)),
			Description: "Ordered DNS search suffixes. Omission or an empty list clears all suffixes."},
	}
}
func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := dnsAttributes()
	attrs["accept_dns_requests_on_lan_interface_ip"] = schema.BoolAttribute{Required: true,
		Description: "Accept DNS requests sent to Socket LAN interface IP."}
	resp.Schema = schema.Schema{Description: "Manages existing account DNS settings via networkConfig.dns. " +
		"Destroy only removes Terraform ownership; remote settings are retained. Requires NCS_DNS_PUBLIC_API.", Attributes: attrs}
}
func (r *siteSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := dnsAttributes()
	attrs["site_id"] = schema.StringAttribute{Required: true, Description: "Existing site's numeric ID. Destroy restores account DNS " +
		"inheritance.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}}
	resp.Schema = schema.Schema{Description: "Manages one site's DNS override via networkConfig.dns. Destroy " +
		"removes override and restores inheritance. Requires NCS_DNS_PUBLIC_API.", Attributes: attrs}
}
