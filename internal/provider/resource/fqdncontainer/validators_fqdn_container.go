package fqdncontainer

import (
	"context"
	"math"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var fqdnPattern = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

func validMember(s string) bool { return len(s) <= 253 && fqdnPattern.MatchString(s) }
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
func knownString(s types.String) bool { return !s.IsNull() && !s.IsUnknown() }
func validateModel(ctx context.Context, p resourceModel, d *diag.Diagnostics) {
	if !knownString(p.Name) ||
		strings.TrimSpace(p.Name.ValueString()) == "" ||
		!knownString(p.Description) ||
		!knownString(p.Mode) {
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
func validURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Fragment == ""
}
func validInterval(v int64) bool { return v > 0 && v <= math.MaxInt32 }

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
				if !v.IsUnknown() &&
					!v.IsNull() &&
					!validMember(v.(types.String).ValueString()) {
					(*d).AddError("Invalid member",
						"Use canonical lowercase FQDN values without spaces, wildcards or trailing slashes.")
				}
			}
		}
	}
}
