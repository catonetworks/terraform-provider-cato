package socketsiteha

import "github.com/hashicorp/terraform-plugin-framework/types"

type SocketSiteHA struct {
	ID            types.String `tfsdk:"id"`
	SiteID        types.String `tfsdk:"site_id"`
	IsCloudRouter types.Bool   `tfsdk:"is_cloud_router"`
}
