package permissioncatalog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

func (d *catalogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Reads resources and supported actions available for defining custom roles.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Description: "Account identifier for this query.", Computed: true},
			"account_id": iam.DataAccountAttribute(),
			"resources": schema.SetNestedAttribute{Description: "Available permission resources.",
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"resource": schema.StringAttribute{Description: "Permission resource identifier.", Computed: true},
					"supported_actions": schema.SetAttribute{Description: "Actions grantable on this resource: VIEW or EDIT.",
						Computed:    true,
						ElementType: types.StringType},
				}}},
		}}
}
