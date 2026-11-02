package dnssitesettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

func (r *siteSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := nc.SchemaDNSAttributes()
	attrs["site"] = schema.SingleNestedAttribute{
		Required:      true,
		Description:   "Existing site, identified by ID or name. Changing the site replaces its managed override.",
		Attributes:    idname.SchemaNameID("Site"),
		PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier(), objectplanmodifier.RequiresReplace()},
	}
	resp.Schema = schema.Schema{
		Description: "Manages one site's DNS override via networkConfig.dns. Destroy restores account DNS inheritance.",
		Attributes:  attrs,
	}
}
