package atrule

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	applicationcontrol2 "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/applicationcontrol"
	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyschedule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// hydrateAppTenantRestrictionRuleStateFromClient maps a policy read rule into Terraform nested objects.
//
//nolint:gocyclo,funlen
func hydrateAppTenantRestrictionRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.AppTenantRestrictionPolicy_Policy_AppTenantRestriction_Policy_Rules_Rule,
) (AppTenantRestrictionRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := AppTenantRestrictionRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.Application = types.ObjectNull(idname.RefAttrTypes)
	out.Headers = types.ListNull(types.ObjectType{AttrTypes: AppTenantRestrictionHeaderAttrTypes})
	out.Source = types.ObjectNull(applicationcontrol2.ApplicationControlSourceAttrTypes)

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())
	if r.GetAction() != nil {
		out.Action = types.StringValue(string(*r.GetAction()))
	}
	if r.GetSeverity() != nil {
		out.Severity = types.StringValue(string(*r.GetSeverity()))
	}

	if app := r.GetApplication(); app != nil && (app.GetID() != "" || app.GetName() != "") {
		appObj, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(app.GetID()),
			"name": types.StringValue(app.GetName()),
		})
		diags.Append(d...)
		out.Application = appObj
	}

	if hdrs := r.GetHeaders(); len(hdrs) > 0 {
		elems := make([]attr.Value, 0, len(hdrs))
		for _, h := range hdrs {
			if h == nil {
				continue
			}
			o, d := types.ObjectValue(AppTenantRestrictionHeaderAttrTypes, map[string]attr.Value{
				"name":  types.StringValue(h.GetName()),
				"value": types.StringValue(h.GetValue()),
			})
			diags.Append(d...)
			elems = append(elems, o)
		}
		lst, d := types.ListValue(types.ObjectType{AttrTypes: AppTenantRestrictionHeaderAttrTypes}, elems)
		diags.Append(d...)
		out.Headers = lst
	}

	sch := r.GetSchedule()
	if sch != nil {
		var ctfObj types.Object
		if ct := sch.GetCustomTimeframeAppTenantRestriction(); ct != nil {
			var dtf diag.Diagnostics
			ctfObj, dtf = types.ObjectValue(policyschedule.CustomTimeframeAttrTypes, map[string]attr.Value{
				"from": types.StringValue(utils.NormalizePolicyTimestampString(ct.GetFrom())),
				"to":   types.StringValue(utils.NormalizePolicyTimestampString(ct.GetTo())),
			})
			diags.Append(dtf...)
		} else {
			ctfObj = types.ObjectNull(policyschedule.CustomTimeframeAttrTypes)
		}

		var crObj types.Object
		if cr := sch.GetCustomRecurringAppTenantRestriction(); cr != nil {
			fromS := ""
			toS := ""
			if cr.GetFrom() != nil {
				fromS = utils.NormalizePolicyRecurringClock(string(*cr.GetFrom()))
			}
			if cr.GetTo() != nil {
				toS = utils.NormalizePolicyRecurringClock(string(*cr.GetTo()))
			}
			days := convert.ParseList(ctx, types.StringType, cr.GetDays(), "rule.schedule.custom_recurring.days")
			var dcr diag.Diagnostics
			crObj, dcr = types.ObjectValue(policyschedule.FromToDaysAttrTypes, map[string]attr.Value{
				"from": types.StringValue(fromS),
				"to":   types.StringValue(toS),
				"days": days,
			})
			diags.Append(dcr...)
		} else {
			crObj = types.ObjectNull(policyschedule.FromToDaysAttrTypes)
		}

		active := ""
		if sch.GetActiveOn() != nil {
			active = string(*sch.GetActiveOn())
		}
		schObj, d := types.ObjectValue(policyschedule.ScheduleAttrTypes, map[string]attr.Value{
			"active_on":        types.StringValue(active),
			"custom_timeframe": ctfObj,
			"custom_recurring": crObj,
		})
		diags.Append(d...)
		out.Schedule = schObj
	}

	src := r.GetSource()
	if src != nil {
		srcAttrs := map[string]attr.Value{
			"country":             oldidname.ParseNameIDList(ctx, src.GetCountry(), "rule.source.country"),
			"ip":                  convert.ParseList(ctx, types.StringType, src.GetIP(), "rule.source.ip"),
			"host":                oldidname.ParseNameIDList(ctx, src.GetHost(), "rule.source.host"),
			"site":                oldidname.ParseNameIDList(ctx, src.GetSite(), "rule.source.site"),
			"subnet":              convert.ParseList(ctx, types.StringType, src.GetSubnet(), "rule.source.subnet"),
			"ip_range":            convert.ParseFromToList(ctx, src.GetIPRange(), "rule.source.ip_range"),
			"global_ip_range":     oldidname.ParseNameIDList(ctx, src.GetGlobalIPRange(), "rule.source.global_ip_range"),
			"network_interface":   oldidname.ParseNameIDList(ctx, src.GetNetworkInterface(), "rule.source.network_interface"),
			"site_network_subnet": oldidname.ParseNameIDList(ctx, src.GetSiteNetworkSubnet(), "rule.source.site_network_subnet"),
			"floating_subnet":     oldidname.ParseNameIDList(ctx, src.GetFloatingSubnet(), "rule.source.floating_subnet"),
			"user":                oldidname.ParseNameIDList(ctx, src.GetUser(), "rule.source.user"),
			"users_group":         oldidname.ParseNameIDList(ctx, src.GetUsersGroup(), "rule.source.users_group"),
			"group":               oldidname.ParseNameIDList(ctx, src.GetGroup(), "rule.source.group"),
			"system_group":        oldidname.ParseNameIDList(ctx, src.GetSystemGroup(), "rule.source.system_group"),
		}
		o, d := types.ObjectValue(applicationcontrol2.ApplicationControlSourceAttrTypes, srcAttrs)
		diags.Append(d...)
		out.Source = o
	}

	return out, diags
}

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	AppTenantRestrictionPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.AppTenantRestrictionPolicy, error)
	PolicyAppTenantRestrictionAddRule(
		ctx context.Context,
		input cato_models.AppTenantRestrictionAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionAddRule, error)
	PolicyAppTenantRestrictionMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionMoveRule, error)
	PolicyAppTenantRestrictionRemoveRule(
		ctx context.Context,
		input cato_models.AppTenantRestrictionRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionRemoveRule, error)
	PolicyAppTenantRestrictionUpdateRule(
		ctx context.Context,
		input cato_models.AppTenantRestrictionUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionUpdateRule, error)
	PolicyAppTenantRestrictionPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error)
}

func (r *appTenantRestrictionRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) AppTenantRestrictionPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.AppTenantRestrictionPolicy, error) {
	return c.resource.client.Catov2.AppTenantRestrictionPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionAddRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionAddRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionMoveRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionRemoveRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionRemoveRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyAppTenantRestrictionUpdateRule(
	ctx context.Context,
	input cato_models.AppTenantRestrictionUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionUpdateRule, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *appTenantRestrictionRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyAppTenantRestrictionPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyAppTenantRestrictionPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, accountID, interceptors...)
}
