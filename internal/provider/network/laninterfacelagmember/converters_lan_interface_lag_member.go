package laninterfacelagmember

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	socketinterface "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/socketinterface"
)

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(plan LagMember) (resultInput cato_models.UpdateSocketInterfaceInput) {
	destType := plan.DestType.ValueString()
	resultInput = cato_models.UpdateSocketInterfaceInput{
		DestType: cato_models.SocketInterfaceDestType(destType),
		Name:     plan.Name.ValueStringPointer(),
	}
	return resultInput
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *LagMember, siteUpdateSocketInterfaceResponse *cato_go_sdk.SiteUpdateSocketInterface) {
	plan.ID = types.StringValue(
		siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID +
			":" +
			string(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID),
	)
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(plan LagMember) (resultInput cato_models.UpdateSocketInterfaceInput) {
	destType := plan.DestType.ValueString()
	resultInput = cato_models.UpdateSocketInterfaceInput{
		DestType: cato_models.SocketInterfaceDestType(destType),
		Name:     plan.Name.ValueStringPointer(),
	}
	return resultInput
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(plan *LagMember, siteUpdateSocketInterfaceResponse *cato_go_sdk.SiteUpdateSocketInterface) {
	plan.ID = types.StringValue(
		siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SiteID +
			":" +
			string(siteUpdateSocketInterfaceResponse.Site.UpdateSocketInterface.SocketInterfaceID),
	)
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state LagMember) (resultInput cato_models.UpdateSocketInterfaceInput) {
	resultInput = cato_models.UpdateSocketInterfaceInput{
		Name:     state.InterfaceID.ValueStringPointer(),
		DestType: "INTERFACE_DISABLED",
	}
	return resultInput
}

// convertHydrateLanInterfaceLagMemberStateValues converts between the Terraform model and SDK values.
func convertHydrateLanInterfaceLagMemberStateValues(
	input *LagMember,
	originalSiteID *string,
	originalInterfaceID *string,
	priorLagMemberFound *bool,
	iface *cato_go_sdk.AccountSnapshot_AccountSnapshot_Sites_InfoSiteSnapshot_Interfaces,
) {
	siteID := *originalSiteID
	defer func() { *originalSiteID = siteID }()

	interfaceID := *originalInterfaceID
	defer func() { *originalInterfaceID = interfaceID }()

	lagMemberFound := *priorLagMemberFound
	defer func() { *priorLagMemberFound = lagMemberFound }()
	input.Name = types.StringPointerValue(iface.Name)
	input.SiteID = types.StringPointerValue(&siteID)
	input.InterfaceID = types.StringPointerValue(&interfaceID)
	input.DestType = types.StringValue(socketinterface.LanLagMemberDestType)
	lagMemberFound = true
}
