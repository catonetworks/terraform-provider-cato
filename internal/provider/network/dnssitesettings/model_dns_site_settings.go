package dnssitesettings

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SiteSettingsModel contains one site's DNS override and site reference.
type SiteSettingsModel struct {
	ID              types.String `tfsdk:"id"`
	Site            types.Object `tfsdk:"site"`
	PrimaryServer   types.String `tfsdk:"primary_server"`
	SecondaryServer types.String `tfsdk:"secondary_server"`
	Suffix          types.List   `tfsdk:"suffix"`
}
