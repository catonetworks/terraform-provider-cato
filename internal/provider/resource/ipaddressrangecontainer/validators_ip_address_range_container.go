package ipaddressrangecontainer

import (
	"context"
	"net/netip"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	containerhelpers "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/container"
)

const rangeEndpoints = 2

func parseMember(s string) (from, to netip.Addr, valid bool) {
	parts := strings.Split(s,
		"-")
	if len(parts) > rangeEndpoints {
		return netip.Addr{},
			netip.Addr{},
			false
	}
	from,
		err := netip.ParseAddr(parts[0])
	if err != nil ||
		!from.Is4() ||
		from.String() != parts[0] {
		return netip.Addr{},
			netip.Addr{},
			false
	}
	to = from
	if len(parts) == rangeEndpoints {
		to,
			err = netip.ParseAddr(parts[1])
		if err != nil ||
			!to.Is4() ||
			to.String() != parts[1] {
			return netip.Addr{},
				netip.Addr{},
				false
		}
	}
	return from,
		to,
		from.Compare(to) <= 0
}
func validMember(s string) bool { _, _, ok := parseMember(s); return ok }
func memberText(value attr.Value) (string,
	bool) {
	o,
		ok := value.(types.Object)
	if !ok ||
		o.IsNull() ||
		o.IsUnknown() {
		return "",
			false
	}
	a := o.Attributes()
	from,
		okF := a["from"].(types.String)
	to,
		okT := a["to"].(types.String)
	if !okF ||
		!okT ||
		!containerhelpers.KnownString(from) ||
		!containerhelpers.KnownString(to) {
		return "",
			false
	}
	s := from.ValueString() + "-" + to.ValueString()
	return s,
		validMember(s)
}
func validID(id types.String,
	d *diag.Diagnostics) bool {
	if id.IsNull() ||
		id.IsUnknown() ||
		strings.TrimSpace(id.ValueString()) == "" {
		d.AddError("Invalid container ID",
			"A known, nonempty container ID is required.")
		return false
	}
	return true
}

func validateModel(ctx context.Context, p resourceModel, d *diag.Diagnostics) {
	if !containerhelpers.KnownString(p.Name) ||
		strings.TrimSpace(p.Name.ValueString()) == "" ||
		!containerhelpers.KnownString(p.Description) ||
		!containerhelpers.KnownString(p.Mode) {
		d.AddError("Invalid container configuration",
			"Name, description and mode must be known; name must not be empty.")
		return
	}
	switch p.Mode.ValueString() {
	case "INLINE":
		if !p.SyncData.IsNull() ||
			!p.FileType.IsNull() {
			d.AddError("Invalid INLINE configuration",
				"INLINE conflicts with sync_data and file_type.")
		}
		inlineValues(ctx,
			p.Values,
			d)
	case "URL":
		if !p.Values.IsNull() {
			d.AddError("Invalid URL configuration", "URL conflicts with values.")
		}
		buildURL(ctx, p, d)
	default:
		d.AddError("Invalid container mode", "Mode must be INLINE or URL.")
	}
}
func (r *containerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Defer expression-derived unknowns until apply, while validating known combinations.
	if p.Mode.IsUnknown() {
		return
	}
	validateInlineConfig(p, &resp.Diagnostics)

	if p.Mode.ValueString() == "URL" {
		if !p.Values.IsNull() {
			resp.Diagnostics.AddError("Invalid URL configuration",
				"URL conflicts with values.")
		}
		if p.FileType.IsNull() ||
			p.SyncData.IsNull() {
			resp.Diagnostics.AddError("Missing sync settings",
				"URL requires file_type and sync_data.")
		}
	}
}

func validateInlineConfig(p resourceModel, d *diag.Diagnostics) {
	if p.Mode.ValueString() == "INLINE" {
		if !p.FileType.IsNull() ||
			!p.SyncData.IsNull() {
			(*d).AddError("Invalid INLINE configuration",
				"INLINE conflicts with file_type and sync_data.")
		}
		if p.Values.IsNull() {
			(*d).AddError("Missing values",
				"INLINE requires values, including [] for an empty container.")
		}
		if !p.Values.IsUnknown() &&
			!p.Values.IsNull() {
			for _, v := range p.Values.Elements() {
				if !memberHasUnknown(v) &&
					!v.IsNull() &&
					!validTFMember(v) {
					(*d).AddError("Invalid member",
						"Use canonical IPv4 addresses in ordered inclusive from/to ranges.")
				}
			}
		}
	}
}

func validTFMember(v attr.Value) bool { _, ok := memberText(v); return ok }

func memberHasUnknown(v attr.Value) bool {
	if v.IsUnknown() {
		return true
	}
	if o, ok := v.(types.Object); ok {
		for _, field := range o.Attributes() {
			if field.IsUnknown() {
				return true
			}
		}
	}
	return false
}
