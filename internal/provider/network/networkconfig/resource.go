package networkconfig

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// Resource holds the account-scoped SDK client shared by standalone DNS/DHCP resources.
type Resource struct{ Client *catoclient.Service }

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected Cato client, got %T", req.ProviderData))
		return
	}
	r.Client = client
}

func IDAttribute() schema.StringAttribute {
	return schema.StringAttribute{Computed: true, Description: "Resource identifier. Collection item IDs are opaque and must be " +
		"preserved unchanged.",
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
}

func Strings(ctx context.Context, value types.List, diags *diag.Diagnostics) []string {
	var result []string
	diags.Append(value.ElementsAs(ctx, &result, false)...)
	if result == nil {
		result = []string{}
	}
	return result
}

func List(ctx context.Context, values []string, diags *diag.Diagnostics) types.List {
	if len(values) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}
	result, d := types.ListValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	return result
}

func RequirePayload(valid bool, diags *diag.Diagnostics) bool {
	if !valid {
		diags.AddError("Invalid network configuration response", "API returned an absent or malformed payload. Terraform state was retained.")
	}
	return valid
}

// IsNotFound checks both structured code and leaf path, never error message text.
func IsNotFound(err error, leaf string) bool {
	var response *clientv2.ErrorResponse
	if errors.As(err, &response) && response.GqlErrors != nil {
		if len(*response.GqlErrors) == 0 {
			return false
		}
		for _, e := range *response.GqlErrors {
			if e.Extensions["code"] != "NOT_FOUND" || len(e.Path) == 0 || fmt.Sprint(e.Path[len(e.Path)-1]) != leaf {
				return false
			}
		}
		return true
	}
	var list *clientv2.GqlErrorList
	if errors.As(err, &list) {
		if len(list.Errors) == 0 {
			return false
		}
		for _, e := range list.Errors {
			if e.Extensions["code"] != "NOT_FOUND" || len(e.Path) == 0 || fmt.Sprint(e.Path[len(e.Path)-1]) != leaf {
				return false
			}
		}
		return true
	}
	return false
}

func APIError(diags *diag.Diagnostics, operation string, err error) {
	diags.AddError("Network configuration API error", operation+": "+err.Error())
}

func ImportAccount(ctx context.Context, r *Resource, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.Client == nil || req.ID != r.Client.AccountId {
		resp.Diagnostics.AddError("Account mismatch", "Import ID must equal provider account_id.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func CheckAccount(r *Resource, id types.String, diags *diag.Diagnostics) bool {
	if r.Client == nil || id.ValueString() != r.Client.AccountId {
		diags.AddError("Account mismatch", "State identity differs from provider account_id. Use the "+
			"original account or import the intended account settings.")
		return false
	}
	return true
}

// IPValidator accepts IP addresses only; unknown expression values are checked at apply.
type IPValidator struct{}

func (IPValidator) Description(context.Context) string               { return "Must be a valid IP address." }
func (v IPValidator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }
func (IPValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := netip.ParseAddr(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid IP address", "Expected an IPv4 or IPv6 address.")
	}
}
