package statichost

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(plan StaticHost) (resultInput cato_models.AddStaticHostInput) {
	resultInput = cato_models.AddStaticHostInput{
		Name:       plan.Name.ValueString(),
		IP:         plan.IP.ValueString(),
		MacAddress: plan.MacAddress.ValueStringPointer(),
	}
	return resultInput
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(state *StaticHost, staticHost *cato_go_sdk.SiteStaticHost) {
	host := staticHost.GetSite().GetStaticHost().GetHost()
	state.ID = types.StringValue(host.HostID)
	state.Name = types.StringValue(host.Name)
	state.IP = types.StringValue(host.IP)
	if host.MacAddress == nil {
		state.MacAddress = types.StringNull()
	} else {
		state.MacAddress = types.StringValue(*host.MacAddress)
	}
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(plan StaticHost) (resultInput cato_models.UpdateStaticHostInput) {
	resultInput = cato_models.UpdateStaticHostInput{
		Name:       plan.Name.ValueStringPointer(),
		IP:         plan.IP.ValueStringPointer(),
		MacAddress: plan.MacAddress.ValueStringPointer(),
	}
	return resultInput
}

func staticHostSiteRefFromTerraform(state StaticHost) cato_models.SiteRefInput {
	return cato_models.SiteRefInput{
		By:    cato_models.ObjectRefByID,
		Input: state.SiteID.ValueString(),
	}
}
