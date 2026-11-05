package terminalserversection

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type parseReadResponse = cato.PolicyTerminalServerPolicy_Policy_TerminalServer_Policy_Sections_Section

func parseRead(ctx context.Context, value *parseReadResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseAddResponse = cato.PolicyTerminalServerAddSection_Policy_TerminalServer_AddSection_Section_Section

func parseAdd(ctx context.Context, value *parseAddResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseUpdateResponse = cato.PolicyTerminalServerUpdateSection_Policy_TerminalServer_UpdateSection_Section_Section

func parseUpdate(ctx context.Context, value *parseUpdateResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type parseMoveResponse = cato.PolicyTerminalServerMoveSection_Policy_TerminalServer_MoveSection_Section_Section

func parseMove(ctx context.Context, value *parseMoveResponse, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing section object.")
		return types.ObjectNull(sectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}
