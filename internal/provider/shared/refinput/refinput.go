// Package refinput converts Terraform {id,name} reference objects into the
// SDK's typed *RefInput values (by-name preferred, by-id fallback), shared by
// the policy-family rule converters.
package refinput

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var asOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

// FromObject converts a single {id,name} object via the provided constructor.
// Returns the zero value and false when the object is null/unknown.
func FromObject[T any](
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
	label string,
	mk func(by cato_models.ObjectRefBy, input string) T,
) (T, bool) {
	var zero T
	if obj.IsNull() || obj.IsUnknown() {
		return zero, false
	}
	ref := idname.Ref{}
	diags.Append(obj.As(ctx, &ref, asOpts)...)
	if diags.HasError() {
		return zero, false
	}
	out, err := utils.TransformObjectRefInput(ref)
	if err != nil {
		diags.AddError("Invalid "+label+" reference", err.Error())
		return zero, false
	}
	return mk(cato_models.ObjectRefBy(out.By), out.Input), true
}

// FromSet converts a set of {id,name} objects via the provided constructor.
// Returns nil when the set is null/unknown.
func FromSet[T any](
	ctx context.Context,
	set types.Set,
	diags *diag.Diagnostics,
	label string,
	mk func(by cato_models.ObjectRefBy, input string) T,
) []T {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var refs []idname.Ref
	diags.Append(set.ElementsAs(ctx, &refs, false)...)
	if diags.HasError() {
		return nil
	}
	out := make([]T, 0, len(refs))
	for _, ref := range refs {
		r, err := utils.TransformObjectRefInput(ref)
		if err != nil {
			diags.AddError("Invalid "+label+" reference", err.Error())
			return nil
		}
		out = append(out, mk(cato_models.ObjectRefBy(r.By), r.Input))
	}
	return out
}

// Constructors for the common SDK ref inputs.

func User(by cato_models.ObjectRefBy, input string) *cato_models.UserRefInput {
	return &cato_models.UserRefInput{By: by, Input: input}
}

func UsersGroup(by cato_models.ObjectRefBy, input string) *cato_models.UsersGroupRefInput {
	return &cato_models.UsersGroupRefInput{By: by, Input: input}
}

func Country(by cato_models.ObjectRefBy, input string) *cato_models.CountryRefInput {
	return &cato_models.CountryRefInput{By: by, Input: input}
}

func DeviceProfile(by cato_models.ObjectRefBy, input string) *cato_models.DeviceProfileRefInput {
	return &cato_models.DeviceProfileRefInput{By: by, Input: input}
}

func GlobalIPRange(by cato_models.ObjectRefBy, input string) *cato_models.GlobalIPRangeRefInput {
	return &cato_models.GlobalIPRangeRefInput{By: by, Input: input}
}

func Application(by cato_models.ObjectRefBy, input string) *cato_models.ApplicationRefInput {
	return &cato_models.ApplicationRefInput{By: cato_models.ObjectRefBy(by), Input: input}
}

func AllocatedIP(by cato_models.ObjectRefBy, input string) *cato_models.AllocatedIPRefInput {
	return &cato_models.AllocatedIPRefInput{By: by, Input: input}
}

func Site(by cato_models.ObjectRefBy, input string) *cato_models.SiteRefInput {
	return &cato_models.SiteRefInput{By: by, Input: input}
}

func SubscriptionMailingList(by cato_models.ObjectRefBy, input string) *cato_models.SubscriptionMailingListRefInput {
	return &cato_models.SubscriptionMailingListRefInput{By: by, Input: input}
}

func SubscriptionGroup(by cato_models.ObjectRefBy, input string) *cato_models.SubscriptionGroupRefInput {
	return &cato_models.SubscriptionGroupRefInput{By: by, Input: input}
}

func SubscriptionWebhook(by cato_models.ObjectRefBy, input string) *cato_models.SubscriptionWebhookRefInput {
	return &cato_models.SubscriptionWebhookRefInput{By: by, Input: input}
}

func Group(by cato_models.ObjectRefBy, input string) *cato_models.GroupRefInput {
	return &cato_models.GroupRefInput{By: by, Input: input}
}

func Host(by cato_models.ObjectRefBy, input string) *cato_models.HostRefInput {
	return &cato_models.HostRefInput{By: by, Input: input}
}

func NetworkInterface(by cato_models.ObjectRefBy, input string) *cato_models.NetworkInterfaceRefInput {
	return &cato_models.NetworkInterfaceRefInput{By: by, Input: input}
}

func SiteNetworkSubnet(by cato_models.ObjectRefBy, input string) *cato_models.SiteNetworkSubnetRefInput {
	return &cato_models.SiteNetworkSubnetRefInput{By: by, Input: input}
}

func FloatingSubnet(by cato_models.ObjectRefBy, input string) *cato_models.FloatingSubnetRefInput {
	return &cato_models.FloatingSubnetRefInput{By: by, Input: input}
}

func CustomApplication(by cato_models.ObjectRefBy, input string) *cato_models.CustomApplicationRefInput {
	return &cato_models.CustomApplicationRefInput{By: by, Input: input}
}

// StringsFromList decodes a types.List of strings (nil when null/unknown).
func StringsFromList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(list.ElementsAs(ctx, &out, false)...)
	return out
}
