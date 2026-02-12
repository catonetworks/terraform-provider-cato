package provider

import (
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type idRefTypes interface {
	cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Country |
		cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Source_UsersGroup |
		cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Source_User |
		cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Applications_Application |
		cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Tracking_Alert_MailingList |
		cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Tracking_Alert_Webhook |
		cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Device
}

func parseIDRef[T idRefTypes](items []*T) (out []IdNameRefModel) {
	type idn struct{ ID, Name string }
	for _, i := range items {
		if i == nil {
			continue
		}
		val := idn(*i)
		out = append(out, IdNameRefModel{
			ID:   types.StringValue(val.ID),
			Name: types.StringValue(val.Name),
		})
	}
	return out
}

func parseStringList[T fmt.Stringer](stringers []T) []types.String {
	if len(stringers) == 0 {
		return nil
	}
	out := make([]types.String, 0, len(stringers))
	for _, o := range stringers {
		out = append(out, types.StringValue(o.String()))
	}
	return out
}

type hasValuer interface {
	IsUnknown() bool
	IsNull() bool
}

func hasValue(v hasValuer) bool { return (!v.IsUnknown()) && (!v.IsNull()) }
