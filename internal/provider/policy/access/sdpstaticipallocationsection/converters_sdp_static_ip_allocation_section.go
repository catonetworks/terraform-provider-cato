package sdpstaticipallocationsection

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type parseReadResponse = cato.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy_Sections_Section

func parseRead(ctx context.Context, value *parseReadResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseAddResponse = cato.PolicySdpStaticIpAllocationAddSection_Policy_SdpStaticIPAllocation_AddSection_Section_Section

func parseAdd(ctx context.Context, value *parseAddResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseUpdateResponse = cato.PolicySdpStaticIpAllocationUpdateSection_Policy_SdpStaticIPAllocation_UpdateSection_Section_Section

func parseUpdate(ctx context.Context, value *parseUpdateResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseMoveResponse = cato.PolicySdpStaticIpAllocationMoveSection_Policy_SdpStaticIPAllocation_MoveSection_Section_Section

func parseMove(ctx context.Context, value *parseMoveResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}
