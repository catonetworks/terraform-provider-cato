package permissioncatalog

import (
	"context"

	sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func catalogState(ctx context.Context, account string, response *sdk.RbacRoleManagementPermissionCatalog) (Model, diag.Diagnostics) {
	var d diag.Diagnostics
	result := Model{ID: types.StringValue(account), AccountID: types.StringValue(account)}
	catalog := response.GetRbac().GetRoleManagement().GetPermissionCatalog()
	if catalog == nil || catalog.Resource == nil {
		d.AddError("Invalid permission catalog", "API returned no complete permission catalog.")
		return result, d
	}
	values := make([]attr.Value, 0, len(catalog.Resource))
	seen := map[string]bool{}
	for _, resource := range catalog.Resource {
		if resource == nil || resource.Resource == "" || resource.SupportedAction == nil || seen[resource.Resource] {
			d.AddError("Invalid permission catalog", "API returned an incomplete or repeated resource.")
			continue
		}
		seen[resource.Resource] = true
		actions := make([]string, 0, len(resource.SupportedAction))
		for _, action := range resource.SupportedAction {
			if action != "VIEW" && action != "EDIT" {
				d.AddError("Invalid permission catalog", "API returned an unsupported action.")
				continue
			}
			actions = append(actions, string(action))
		}
		set, diags := types.SetValueFrom(ctx, types.StringType, actions)
		d.Append(diags...)
		object,
			diags := types.ObjectValue(catalogAttrTypes,
			map[string]attr.Value{"resource": types.StringValue(resource.Resource),
				"supported_actions": set})
		d.Append(diags...)
		values = append(values, object)
	}
	var diags diag.Diagnostics
	result.Resources, diags = types.SetValue(catalogObjectType, values)
	d.Append(diags...)
	return result, d
}
