package statichost

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// resolveSiteID resolves the host's parent site to its numeric site ID and
// validates that the site exists, via entityLookup. Exactly one of site_name /
// site_id is set in config (enforced by the schema's ExactlyOneOf validator).
//
//   - by name: entityLookup(type=site, search=name). The `search` argument is a
//     fuzzy match, so results are filtered to an EXACT name match; site names are
//     unique, so exactly one must match (0 -> not found, >1 -> ambiguous).
//   - by id:   entityLookup(type=site, entityIDs=[id]) must return that one site.
//
// The returned id is used in the add/update API call and written back to state,
// so after the first apply state always carries the concrete site_id regardless
// of whether the config used site_name or site_id.
func (r *staticHostResource) resolveSiteID(ctx context.Context, model StaticHost) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := r.resourceAPI()

	if !model.SiteName.IsNull() && model.SiteName.ValueString() != "" {
		name := model.SiteName.ValueString()
		res, err := api.EntityLookup(
			ctx,
			r.client.AccountId,
			cato_models.EntityTypeSite,
			nil,   // limit
			nil,   // from
			nil,   // parent
			&name, // search (fuzzy -- filtered to exact match below)
			nil,   // entityIDs
			nil,   // sort
			nil,   // filters
			nil,   // helperFields
		)
		if err != nil {
			diags.AddError("Catov2 API EntityLookup error", err.Error())
			return "", diags
		}

		matchID := ""
		matchCount := 0
		for _, item := range res.EntityLookup.GetItems() {
			if item == nil {
				continue
			}
			entity := item.GetEntity()
			entityName := entity.GetName()
			if entityName != nil && *entityName == name {
				matchCount++
				matchID = entity.GetID()
			}
		}
		if matchCount == 0 {
			diags.AddError(
				"Site not found",
				fmt.Sprintf("No site named %q exists in account %s.", name, r.client.AccountId),
			)
			return "", diags
		}
		if matchCount > 1 {
			diags.AddError(
				"Ambiguous site name",
				fmt.Sprintf("%d sites named %q exist in account %s; site names are expected to be unique.",
					matchCount, name, r.client.AccountId),
			)
			return "", diags
		}
		return matchID, diags
	}

	// site_id path: reject unresolved configuration before making an API call.
	if model.SiteID.IsNull() || model.SiteID.IsUnknown() || model.SiteID.ValueString() == "" {
		diags.AddError("Missing or unknown site reference", "Set a known, non-empty site_id or site_name before creating a static host.")
		return "", diags
	}
	id := model.SiteID.ValueString()
	res, err := api.EntityLookup(
		ctx,
		r.client.AccountId,
		cato_models.EntityTypeSite,
		nil,          // limit
		nil,          // from
		nil,          // parent
		nil,          // search
		[]string{id}, // entityIDs
		nil,          // sort
		nil,          // filters
		nil,          // helperFields
	)
	if err != nil {
		diags.AddError("Catov2 API EntityLookup error", err.Error())
		return "", diags
	}
	if len(res.EntityLookup.GetItems()) != 1 {
		diags.AddError(
			"Site not found",
			fmt.Sprintf("No site with ID %q exists in account %s.", id, r.client.AccountId),
		)
		return "", diags
	}
	return id, diags
}
