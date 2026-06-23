package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// hydrateApplicationControlRuleStateFromClient maps an Application Control policy rule into Terraform nested objects.
func hydrateApplicationControlRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule,
) (ApplicationControlRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := ApplicationControlRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())
	out.RuleType = types.StringValue(string(r.RuleType))

	out.ApplicationRule = types.ObjectNull(applicationControlTypedRuleAttrTypes)
	out.DataRule = types.ObjectNull(applicationControlTypedRuleAttrTypes)
	out.FileRule = types.ObjectNull(applicationControlTypedRuleAttrTypes)

	switch r.RuleType {
	case cato_models.ApplicationControlRuleTypeApplication:
		if ar := r.GetApplicationRule(); ar != nil {
			obj, d := applicationControlTypedRuleStateFromApplicationRule(ctx, ar)
			diags.Append(d...)
			out.ApplicationRule = obj
		}
	case cato_models.ApplicationControlRuleTypeData:
		if dr := r.GetDataRule(); dr != nil {
			obj, d := applicationControlTypedRuleStateFromDataRule(ctx, dr)
			diags.Append(d...)
			out.DataRule = obj
		}
	case cato_models.ApplicationControlRuleTypeFile:
		if fr := r.GetFileRule(); fr != nil {
			obj, d := applicationControlTypedRuleStateFromFileRule(ctx, fr)
			diags.Append(d...)
			out.FileRule = obj
		}
	default:
		diags.AddError("application control rule", fmt.Sprintf("unsupported rule_type %q", r.RuleType))
	}

	return out, diags
}

func applicationControlTypedRuleStateFromApplicationRule(
	ctx context.Context,
	ar *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ar == nil {
		return types.ObjectNull(applicationControlTypedRuleAttrTypes), diags
	}

	sch := ar.GetSchedule()
	schObj := acScheduleObjectFromApplicationRuleSchedule(ctx, sch, &diags)

	src := ar.GetSource()
	srcObj := acSourceObjectFromApplicationRuleSource(ctx, src, &diags)

	tr := ar.GetTracking()
	trObj := acTrackingObjectFromApplicationRuleTracking(ctx, tr, &diags)

	appObj := wanApplicationObjectFromApplicationRuleApplication(ctx, ar.GetApplication(), &diags)

	acList := acAccessMethodListFromApplicationRule(ctx, ar.GetAccessMethod(), &diags)
	activityList := acApplicationActivityListFromApplicationRule(ar.GetApplicationActivity(), &diags)
	actSatisfy := types.StringValue("ALL")
	if s := ar.GetApplicationActivitySatisfy(); s != nil {
		actSatisfy = types.StringValue(string(*s))
	}

	actionCfg := types.ObjectNull(applicationControlActionConfigAttrTypes)
	if cfg := ar.GetActionConfig(); cfg != nil {
		un := parseNameIDList(ctx, cfg.GetUserNotification(), "application_rule.action_config.user_notification")
		o, d := types.ObjectValue(applicationControlActionConfigAttrTypes, map[string]attr.Value{
			"user_notification": un,
		})
		diags.Append(d...)
		actionCfg = o
	}

	attrs := map[string]attr.Value{
		"action":                       types.StringValue(ar.GetAction().String()),
		"severity":                     types.StringValue(ar.GetSeverity().String()),
		"schedule":                     schObj,
		"source":                       srcObj,
		"tracking":                     trObj,
		"device":                       parseNameIDList(ctx, ar.GetDevice(), "application_rule.device"),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
		"action_config":                actionCfg,
		"file_attribute":               types.ListNull(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}),
		"file_attribute_satisfy":       types.StringNull(),
		"dlp_profile":                  types.ObjectNull(applicationControlDlpProfileAttrTypes),
	}
	o, d := types.ObjectValue(applicationControlTypedRuleAttrTypes, attrs)
	diags.Append(d...)
	return o, diags
}

func applicationControlTypedRuleStateFromDataRule(
	ctx context.Context,
	dr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if dr == nil {
		return types.ObjectNull(applicationControlTypedRuleAttrTypes), diags
	}

	schObj := acScheduleObjectFromDataRuleSchedule(ctx, dr.GetSchedule(), &diags)
	srcObj := acSourceObjectFromDataRuleSource(ctx, dr.GetSource(), &diags)
	trObj := acTrackingObjectFromDataRuleTracking(ctx, dr.GetTracking(), &diags)
	appObj := wanApplicationObjectFromDataRuleApplication(ctx, dr.GetApplication(), &diags)
	acList := acAccessMethodListFromDataRule(ctx, dr.GetAccessMethod(), &diags)
	activityList := acApplicationActivityListFromDataRule(dr.GetApplicationActivity(), &diags)
	actSatisfy := types.StringValue("ALL")
	if s := dr.GetApplicationActivitySatisfy(); s != nil {
		actSatisfy = types.StringValue(string(*s))
	}

	actionCfg := types.ObjectNull(applicationControlActionConfigAttrTypes)
	if cfg := dr.GetActionConfig(); cfg != nil {
		un := parseNameIDList(ctx, cfg.GetUserNotification(), "data_rule.action_config.user_notification")
		o, d := types.ObjectValue(applicationControlActionConfigAttrTypes, map[string]attr.Value{
			"user_notification": un,
		})
		diags.Append(d...)
		actionCfg = o
	}

	dlpObj := types.ObjectNull(applicationControlDlpProfileAttrTypes)
	if dp := dr.GetDlpProfile(); dp != nil {
		o, d := types.ObjectValue(applicationControlDlpProfileAttrTypes, map[string]attr.Value{
			"content_profile": parseNameIDList(ctx, dp.GetContentProfile(), "data_rule.dlp_profile.content_profile"),
			"edm_profile":     parseNameIDList(ctx, dp.GetEdmProfile(), "data_rule.dlp_profile.edm_profile"),
		})
		diags.Append(d...)
		dlpObj = o
	}

	faList := acFileAttributeListFromDataRule(ctx, dr.GetFileAttribute(), &diags)

	attrs := map[string]attr.Value{
		"action":                       types.StringValue(dr.GetAction().String()),
		"severity":                     types.StringValue(dr.GetSeverity().String()),
		"schedule":                     schObj,
		"source":                       srcObj,
		"tracking":                     trObj,
		"device":                       parseNameIDList(ctx, dr.GetDevice(), "data_rule.device"),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
		"action_config":                actionCfg,
		"file_attribute":               faList,
		"file_attribute_satisfy":       types.StringValue(dr.GetFileAttributeSatisfy().String()),
		"dlp_profile":                  dlpObj,
	}
	o, d := types.ObjectValue(applicationControlTypedRuleAttrTypes, attrs)
	diags.Append(d...)
	return o, diags
}

func applicationControlTypedRuleStateFromFileRule(
	ctx context.Context,
	fr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule,
) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if fr == nil {
		return types.ObjectNull(applicationControlTypedRuleAttrTypes), diags
	}

	schObj := acScheduleObjectFromFileRuleSchedule(ctx, fr.GetSchedule(), &diags)
	srcObj := acSourceObjectFromFileRuleSource(ctx, fr.GetSource(), &diags)
	trObj := acTrackingObjectFromFileRuleTracking(ctx, fr.GetTracking(), &diags)
	appObj := wanApplicationObjectFromFileRuleApplication(ctx, fr.GetApplication(), &diags)
	acList := acAccessMethodListFromFileRule(ctx, fr.GetAccessMethod(), &diags)
	activityList := acApplicationActivityListFromFileRule(fr.GetApplicationActivity(), &diags)
	actSatisfy := types.StringValue("ALL")
	if s := fr.GetApplicationActivitySatisfy(); s != nil {
		actSatisfy = types.StringValue(string(*s))
	}

	actionCfg := types.ObjectNull(applicationControlActionConfigAttrTypes)
	if cfg := fr.GetActionConfig(); cfg != nil {
		un := parseNameIDList(ctx, cfg.GetUserNotification(), "file_rule.action_config.user_notification")
		o, d := types.ObjectValue(applicationControlActionConfigAttrTypes, map[string]attr.Value{
			"user_notification": un,
		})
		diags.Append(d...)
		actionCfg = o
	}

	faList := acFileAttributeListFromFileRule(ctx, fr.GetFileAttribute(), &diags)

	attrs := map[string]attr.Value{
		"action":                       types.StringValue(fr.GetAction().String()),
		"severity":                     types.StringValue(fr.GetSeverity().String()),
		"schedule":                     schObj,
		"source":                       srcObj,
		"tracking":                     trObj,
		"device":                       parseNameIDList(ctx, fr.GetDevice(), "file_rule.device"),
		"access_method":                acList,
		"application":                  appObj,
		"application_activity":         activityList,
		"application_activity_satisfy": actSatisfy,
		"action_config":                actionCfg,
		"file_attribute":               faList,
		"file_attribute_satisfy":       types.StringValue(fr.GetFileAttributeSatisfy().String()),
		"dlp_profile":                  types.ObjectNull(applicationControlDlpProfileAttrTypes),
	}
	o, d := types.ObjectValue(applicationControlTypedRuleAttrTypes, attrs)
	diags.Append(d...)
	return o, diags
}

func acScheduleObjectFromApplicationRuleSchedule(
	ctx context.Context,
	sch *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Schedule,
	diags *diag.Diagnostics,
) types.Object {
	if sch == nil {
		return types.ObjectNull(ScheduleAttrTypes)
	}
	active := ""
	if sch.GetActiveOn() != nil {
		active = sch.GetActiveOn().String()
	}
	var ctfObj types.Object
	if ct := sch.GetCustomTimeframeApplicationRule(); ct != nil {
		o, d := types.ObjectValue(FromToAttrTypes, map[string]attr.Value{
			"from": types.StringValue(normalizePolicyTimestampString(ct.GetFrom())),
			"to":   types.StringValue(normalizePolicyTimestampString(ct.GetTo())),
		})
		diags.Append(d...)
		ctfObj = o
	} else {
		ctfObj = types.ObjectNull(FromToAttrTypes)
	}
	var crObj types.Object
	if cr := sch.GetCustomRecurringApplicationRule(); cr != nil {
		dayStrings := make([]string, 0, len(cr.GetDays()))
		for _, d := range cr.GetDays() {
			dayStrings = append(dayStrings, d.String())
		}
		days := parseList(ctx, types.StringType, dayStrings, "application_rule.schedule.custom_recurring.days")
		o, d := types.ObjectValue(FromToDaysAttrTypes, map[string]attr.Value{
			"from": types.StringValue(normalizePolicyRecurringClock(string(cr.From))),
			"to":   types.StringValue(normalizePolicyRecurringClock(string(cr.To))),
			"days": days,
		})
		diags.Append(d...)
		if d.HasError() || o.IsUnknown() {
			crObj = types.ObjectNull(FromToDaysAttrTypes)
		} else {
			crObj = o
		}
	} else {
		crObj = types.ObjectNull(FromToDaysAttrTypes)
	}
	o, d := types.ObjectValue(ScheduleAttrTypes, map[string]attr.Value{
		"active_on":        types.StringValue(active),
		"custom_timeframe": ctfObj,
		"custom_recurring": crObj,
	})
	diags.Append(d...)
	if d.HasError() || o.IsUnknown() {
		return types.ObjectNull(ScheduleAttrTypes)
	}
	return o
}

func acScheduleObjectFromDataRuleSchedule(
	ctx context.Context,
	sch *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Schedule,
	diags *diag.Diagnostics,
) types.Object {
	if sch == nil {
		return types.ObjectNull(ScheduleAttrTypes)
	}
	active := ""
	if sch.GetActiveOn() != nil {
		active = sch.GetActiveOn().String()
	}
	var ctfObj types.Object
	if ct := sch.GetCustomTimeframeDataRule(); ct != nil {
		o, d := types.ObjectValue(FromToAttrTypes, map[string]attr.Value{
			"from": types.StringValue(normalizePolicyTimestampString(ct.GetFrom())),
			"to":   types.StringValue(normalizePolicyTimestampString(ct.GetTo())),
		})
		diags.Append(d...)
		ctfObj = o
	} else {
		ctfObj = types.ObjectNull(FromToAttrTypes)
	}
	var crObj types.Object
	if cr := sch.GetCustomRecurringDataRule(); cr != nil {
		dayStrings := make([]string, 0, len(cr.GetDays()))
		for _, d := range cr.GetDays() {
			dayStrings = append(dayStrings, d.String())
		}
		days := parseList(ctx, types.StringType, dayStrings, "data_rule.schedule.custom_recurring.days")
		fromS := ""
		toS := ""
		if cr.GetFrom() != nil {
			fromS = normalizePolicyRecurringClock(string(*cr.GetFrom()))
		}
		if cr.GetTo() != nil {
			toS = normalizePolicyRecurringClock(string(*cr.GetTo()))
		}
		o, d := types.ObjectValue(FromToDaysAttrTypes, map[string]attr.Value{
			"from": types.StringValue(fromS),
			"to":   types.StringValue(toS),
			"days": days,
		})
		diags.Append(d...)
		if d.HasError() || o.IsUnknown() {
			crObj = types.ObjectNull(FromToDaysAttrTypes)
		} else {
			crObj = o
		}
	} else {
		crObj = types.ObjectNull(FromToDaysAttrTypes)
	}
	o, d := types.ObjectValue(ScheduleAttrTypes, map[string]attr.Value{
		"active_on":        types.StringValue(active),
		"custom_timeframe": ctfObj,
		"custom_recurring": crObj,
	})
	diags.Append(d...)
	if d.HasError() || o.IsUnknown() {
		return types.ObjectNull(ScheduleAttrTypes)
	}
	return o
}

func acScheduleObjectFromFileRuleSchedule(
	ctx context.Context,
	sch *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Schedule,
	diags *diag.Diagnostics,
) types.Object {
	if sch == nil {
		return types.ObjectNull(ScheduleAttrTypes)
	}
	active := ""
	if sch.GetActiveOn() != nil {
		active = sch.GetActiveOn().String()
	}
	var ctfObj types.Object
	if ct := sch.GetCustomTimeframeFileRule(); ct != nil {
		o, d := types.ObjectValue(FromToAttrTypes, map[string]attr.Value{
			"from": types.StringValue(normalizePolicyTimestampString(ct.GetFrom())),
			"to":   types.StringValue(normalizePolicyTimestampString(ct.GetTo())),
		})
		diags.Append(d...)
		ctfObj = o
	} else {
		ctfObj = types.ObjectNull(FromToAttrTypes)
	}
	var crObj types.Object
	if cr := sch.GetCustomRecurringFileRule(); cr != nil {
		dayStrings := make([]string, 0, len(cr.GetDays()))
		for _, d := range cr.GetDays() {
			dayStrings = append(dayStrings, d.String())
		}
		days := parseList(ctx, types.StringType, dayStrings, "file_rule.schedule.custom_recurring.days")
		fromS := ""
		toS := ""
		if cr.GetFrom() != nil {
			fromS = normalizePolicyRecurringClock(string(*cr.GetFrom()))
		}
		if cr.GetTo() != nil {
			toS = normalizePolicyRecurringClock(string(*cr.GetTo()))
		}
		o, d := types.ObjectValue(FromToDaysAttrTypes, map[string]attr.Value{
			"from": types.StringValue(fromS),
			"to":   types.StringValue(toS),
			"days": days,
		})
		diags.Append(d...)
		if d.HasError() || o.IsUnknown() {
			crObj = types.ObjectNull(FromToDaysAttrTypes)
		} else {
			crObj = o
		}
	} else {
		crObj = types.ObjectNull(FromToDaysAttrTypes)
	}
	o, d := types.ObjectValue(ScheduleAttrTypes, map[string]attr.Value{
		"active_on":        types.StringValue(active),
		"custom_timeframe": ctfObj,
		"custom_recurring": crObj,
	})
	diags.Append(d...)
	if d.HasError() || o.IsUnknown() {
		return types.ObjectNull(ScheduleAttrTypes)
	}
	return o
}

func acSourceObjectFromApplicationRuleSource(
	ctx context.Context,
	src *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Source,
	diags *diag.Diagnostics,
) types.Object {
	if src == nil {
		return types.ObjectNull(ApplicationControlSourceAttrTypes)
	}
	srcAttrs := map[string]attr.Value{
		"country":             parseNameIDList(ctx, src.GetCountry(), "application_rule.source.country"),
		"ip":                  parseList(ctx, types.StringType, src.GetIP(), "application_rule.source.ip"),
		"host":                parseNameIDList(ctx, src.GetHost(), "application_rule.source.host"),
		"site":                parseNameIDList(ctx, src.GetSite(), "application_rule.source.site"),
		"subnet":              parseList(ctx, types.StringType, src.GetSubnet(), "application_rule.source.subnet"),
		"ip_range":            parseFromToList(ctx, src.GetIPRange(), "application_rule.source.ip_range"),
		"global_ip_range":     parseNameIDList(ctx, src.GetGlobalIPRange(), "application_rule.source.global_ip_range"),
		"network_interface":   parseNameIDList(ctx, src.GetNetworkInterface(), "application_rule.source.network_interface"),
		"site_network_subnet": parseNameIDList(ctx, src.GetSiteNetworkSubnet(), "application_rule.source.site_network_subnet"),
		"floating_subnet":     parseNameIDList(ctx, src.GetFloatingSubnet(), "application_rule.source.floating_subnet"),
		"user":                parseNameIDList(ctx, src.GetUser(), "application_rule.source.user"),
		"users_group":         parseNameIDList(ctx, src.GetUsersGroup(), "application_rule.source.users_group"),
		"group":               parseNameIDList(ctx, src.GetGroup(), "application_rule.source.group"),
		"system_group":        parseNameIDList(ctx, src.GetSystemGroup(), "application_rule.source.system_group"),
	}
	o, d := types.ObjectValue(ApplicationControlSourceAttrTypes, srcAttrs)
	diags.Append(d...)
	return o
}

func acSourceObjectFromDataRuleSource(
	ctx context.Context,
	src *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Source,
	diags *diag.Diagnostics,
) types.Object {
	if src == nil {
		return types.ObjectNull(ApplicationControlSourceAttrTypes)
	}
	srcAttrs := map[string]attr.Value{
		"country":             parseNameIDList(ctx, src.GetCountry(), "data_rule.source.country"),
		"ip":                  parseList(ctx, types.StringType, src.GetIP(), "data_rule.source.ip"),
		"host":                parseNameIDList(ctx, src.GetHost(), "data_rule.source.host"),
		"site":                parseNameIDList(ctx, src.GetSite(), "data_rule.source.site"),
		"subnet":              parseList(ctx, types.StringType, src.GetSubnet(), "data_rule.source.subnet"),
		"ip_range":            parseFromToList(ctx, src.GetIPRange(), "data_rule.source.ip_range"),
		"global_ip_range":     parseNameIDList(ctx, src.GetGlobalIPRange(), "data_rule.source.global_ip_range"),
		"network_interface":   parseNameIDList(ctx, src.GetNetworkInterface(), "data_rule.source.network_interface"),
		"site_network_subnet": parseNameIDList(ctx, src.GetSiteNetworkSubnet(), "data_rule.source.site_network_subnet"),
		"floating_subnet":     parseNameIDList(ctx, src.GetFloatingSubnet(), "data_rule.source.floating_subnet"),
		"user":                parseNameIDList(ctx, src.GetUser(), "data_rule.source.user"),
		"users_group":         parseNameIDList(ctx, src.GetUsersGroup(), "data_rule.source.users_group"),
		"group":               parseNameIDList(ctx, src.GetGroup(), "data_rule.source.group"),
		"system_group":        parseNameIDList(ctx, src.GetSystemGroup(), "data_rule.source.system_group"),
	}
	o, d := types.ObjectValue(ApplicationControlSourceAttrTypes, srcAttrs)
	diags.Append(d...)
	return o
}

func acSourceObjectFromFileRuleSource(
	ctx context.Context,
	src *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Source,
	diags *diag.Diagnostics,
) types.Object {
	if src == nil {
		return types.ObjectNull(ApplicationControlSourceAttrTypes)
	}
	srcAttrs := map[string]attr.Value{
		"country":             parseNameIDList(ctx, src.GetCountry(), "file_rule.source.country"),
		"ip":                  parseList(ctx, types.StringType, src.GetIP(), "file_rule.source.ip"),
		"host":                parseNameIDList(ctx, src.GetHost(), "file_rule.source.host"),
		"site":                parseNameIDList(ctx, src.GetSite(), "file_rule.source.site"),
		"subnet":              parseList(ctx, types.StringType, src.GetSubnet(), "file_rule.source.subnet"),
		"ip_range":            parseFromToList(ctx, src.GetIPRange(), "file_rule.source.ip_range"),
		"global_ip_range":     parseNameIDList(ctx, src.GetGlobalIPRange(), "file_rule.source.global_ip_range"),
		"network_interface":   parseNameIDList(ctx, src.GetNetworkInterface(), "file_rule.source.network_interface"),
		"site_network_subnet": parseNameIDList(ctx, src.GetSiteNetworkSubnet(), "file_rule.source.site_network_subnet"),
		"floating_subnet":     parseNameIDList(ctx, src.GetFloatingSubnet(), "file_rule.source.floating_subnet"),
		"user":                parseNameIDList(ctx, src.GetUser(), "file_rule.source.user"),
		"users_group":         parseNameIDList(ctx, src.GetUsersGroup(), "file_rule.source.users_group"),
		"group":               parseNameIDList(ctx, src.GetGroup(), "file_rule.source.group"),
		"system_group":        parseNameIDList(ctx, src.GetSystemGroup(), "file_rule.source.system_group"),
	}
	o, d := types.ObjectValue(ApplicationControlSourceAttrTypes, srcAttrs)
	diags.Append(d...)
	return o
}

func acTrackingObjectFromApplicationRuleTracking(
	ctx context.Context,
	tr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Tracking,
	diags *diag.Diagnostics,
) types.Object {
	if tr == nil {
		return types.ObjectNull(TrackingAttrTypes)
	}
	ev := tr.GetEvent()
	evObj, d := types.ObjectValue(TrackingEventAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(ev.GetEnabled()),
	})
	diags.Append(d...)
	al := tr.GetAlert()
	freq := ""
	if al.GetFrequency() != nil {
		freq = al.GetFrequency().String()
	}
	alObj, d := types.ObjectValue(TrackingAlertAttrTypes, map[string]attr.Value{
		"enabled":            types.BoolValue(al.GetEnabled()),
		"frequency":          types.StringValue(freq),
		"subscription_group": parseNameIDListOrEmptySet(ctx, al.GetSubscriptionGroup(), "application_rule.tracking.alert.subscription_group"),
		"webhook":            parseNameIDListOrEmptySet(ctx, al.GetWebhook(), "application_rule.tracking.alert.webhook"),
		"mailing_list":       parseNameIDListOrEmptySet(ctx, al.GetMailingList(), "application_rule.tracking.alert.mailing_list"),
	})
	diags.Append(d...)
	o, d := types.ObjectValue(TrackingAttrTypes, map[string]attr.Value{
		"event": evObj,
		"alert": alObj,
	})
	diags.Append(d...)
	return o
}

func acTrackingObjectFromDataRuleTracking(
	ctx context.Context,
	tr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Tracking,
	diags *diag.Diagnostics,
) types.Object {
	if tr == nil {
		return types.ObjectNull(TrackingAttrTypes)
	}
	ev := tr.GetEvent()
	evObj, d := types.ObjectValue(TrackingEventAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(ev.GetEnabled()),
	})
	diags.Append(d...)
	al := tr.GetAlert()
	freq := ""
	if al.GetFrequency() != nil {
		freq = al.GetFrequency().String()
	}
	alObj, d := types.ObjectValue(TrackingAlertAttrTypes, map[string]attr.Value{
		"enabled":            types.BoolValue(al.GetEnabled()),
		"frequency":          types.StringValue(freq),
		"subscription_group": parseNameIDListOrEmptySet(ctx, al.GetSubscriptionGroup(), "data_rule.tracking.alert.subscription_group"),
		"webhook":            parseNameIDListOrEmptySet(ctx, al.GetWebhook(), "data_rule.tracking.alert.webhook"),
		"mailing_list":       parseNameIDListOrEmptySet(ctx, al.GetMailingList(), "data_rule.tracking.alert.mailing_list"),
	})
	diags.Append(d...)
	o, d := types.ObjectValue(TrackingAttrTypes, map[string]attr.Value{
		"event": evObj,
		"alert": alObj,
	})
	diags.Append(d...)
	return o
}

func acTrackingObjectFromFileRuleTracking(
	ctx context.Context,
	tr *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Tracking,
	diags *diag.Diagnostics,
) types.Object {
	if tr == nil {
		return types.ObjectNull(TrackingAttrTypes)
	}
	ev := tr.GetEvent()
	evObj, d := types.ObjectValue(TrackingEventAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(ev.GetEnabled()),
	})
	diags.Append(d...)
	al := tr.GetAlert()
	freq := ""
	if al.GetFrequency() != nil {
		freq = al.GetFrequency().String()
	}
	alObj, d := types.ObjectValue(TrackingAlertAttrTypes, map[string]attr.Value{
		"enabled":            types.BoolValue(al.GetEnabled()),
		"frequency":          types.StringValue(freq),
		"subscription_group": parseNameIDListOrEmptySet(ctx, al.GetSubscriptionGroup(), "file_rule.tracking.alert.subscription_group"),
		"webhook":            parseNameIDListOrEmptySet(ctx, al.GetWebhook(), "file_rule.tracking.alert.webhook"),
		"mailing_list":       parseNameIDListOrEmptySet(ctx, al.GetMailingList(), "file_rule.tracking.alert.mailing_list"),
	})
	diags.Append(d...)
	o, d := types.ObjectValue(TrackingAttrTypes, map[string]attr.Value{
		"event": evObj,
		"alert": alObj,
	})
	diags.Append(d...)
	return o
}

//nolint:lll
func wanApplicationObjectFromApplicationRuleApplication(
	ctx context.Context,
	a *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application,
	diags *diag.Diagnostics,
) types.Object {
	if a == nil {
		return types.ObjectNull(WanApplicationAttrTypes)
	}
	var apps []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_Application
	if p := a.GetApplication(); p != nil {
		apps = append(apps, p)
	}
	var cats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_AppCategory
	if p := a.GetAppCategory(); p != nil {
		cats = append(cats, p)
	}
	var customs []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_CustomApp
	if p := a.GetCustomApp(); p != nil {
		customs = append(customs, p)
	}
	var customCats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_CustomCategory
	if p := a.GetCustomCategory(); p != nil {
		customCats = append(customCats, p)
	}
	var sac []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_Application_SanctionedAppsCategory
	if p := a.GetSanctionedAppsCategory(); p != nil {
		sac = append(sac, p)
	}
	emptyStrList, _ := types.ListValue(types.StringType, []attr.Value{})
	emptyIPRangeList, _ := types.ListValue(FromToObjectType, []attr.Value{})
	emptyNameIDSet, _ := types.SetValue(NameIDObjectType, []attr.Value{})
	attrs := map[string]attr.Value{
		"application":              parseNameIDListOrEmptySet(ctx, apps, "application_rule.application.application"),
		"custom_app":               parseNameIDListOrEmptySet(ctx, customs, "application_rule.application.custom_app"),
		"app_category":             parseNameIDListOrEmptySet(ctx, cats, "application_rule.application.app_category"),
		"custom_category":          parseNameIDListOrEmptySet(ctx, customCats, "application_rule.application.custom_category"),
		"sanctioned_apps_category": parseNameIDListOrEmptySet(ctx, sac, "application_rule.application.sanctioned_apps_category"),
		"domain":                   emptyStrList,
		"fqdn":                     emptyStrList,
		"ip":                       emptyStrList,
		"subnet":                   emptyStrList,
		"ip_range":                 emptyIPRangeList,
		"global_ip_range":          emptyNameIDSet,
	}
	o, d := types.ObjectValue(WanApplicationAttrTypes, attrs)
	diags.Append(d...)
	return o
}

func wanApplicationObjectFromDataRuleApplication(
	ctx context.Context,
	a *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application,
	diags *diag.Diagnostics,
) types.Object {
	if a == nil {
		return types.ObjectNull(WanApplicationAttrTypes)
	}
	var apps []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_Application
	if p := a.GetApplication(); p != nil {
		apps = append(apps, p)
	}
	var cats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_AppCategory
	if p := a.GetAppCategory(); p != nil {
		cats = append(cats, p)
	}
	var customs []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_CustomApp
	if p := a.GetCustomApp(); p != nil {
		customs = append(customs, p)
	}
	var customCats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_CustomCategory
	if p := a.GetCustomCategory(); p != nil {
		customCats = append(customCats, p)
	}
	var sac []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_Application_SanctionedAppsCategory
	if p := a.GetSanctionedAppsCategory(); p != nil {
		sac = append(sac, p)
	}
	emptyStrList2, _ := types.ListValue(types.StringType, []attr.Value{})
	emptyIPRangeList2, _ := types.ListValue(FromToObjectType, []attr.Value{})
	emptyNameIDSet2, _ := types.SetValue(NameIDObjectType, []attr.Value{})
	attrs := map[string]attr.Value{
		"application":              parseNameIDListOrEmptySet(ctx, apps, "data_rule.application.application"),
		"custom_app":               parseNameIDListOrEmptySet(ctx, customs, "data_rule.application.custom_app"),
		"app_category":             parseNameIDListOrEmptySet(ctx, cats, "data_rule.application.app_category"),
		"custom_category":          parseNameIDListOrEmptySet(ctx, customCats, "data_rule.application.custom_category"),
		"sanctioned_apps_category": parseNameIDListOrEmptySet(ctx, sac, "data_rule.application.sanctioned_apps_category"),
		"domain":                   emptyStrList2,
		"fqdn":                     emptyStrList2,
		"ip":                       emptyStrList2,
		"subnet":                   emptyStrList2,
		"ip_range":                 emptyIPRangeList2,
		"global_ip_range":          emptyNameIDSet2,
	}
	o, d := types.ObjectValue(WanApplicationAttrTypes, attrs)
	diags.Append(d...)
	return o
}

func wanApplicationObjectFromFileRuleApplication(
	ctx context.Context,
	a *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application,
	diags *diag.Diagnostics,
) types.Object {
	if a == nil {
		return types.ObjectNull(WanApplicationAttrTypes)
	}
	var apps []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_Application
	if p := a.GetApplication(); p != nil {
		apps = append(apps, p)
	}
	var cats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_AppCategory
	if p := a.GetAppCategory(); p != nil {
		cats = append(cats, p)
	}
	var customs []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_CustomApp
	if p := a.GetCustomApp(); p != nil {
		customs = append(customs, p)
	}
	var customCats []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_CustomCategory
	if p := a.GetCustomCategory(); p != nil {
		customCats = append(customCats, p)
	}
	var sac []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_Application_SanctionedAppsCategory
	if p := a.GetSanctionedAppsCategory(); p != nil {
		sac = append(sac, p)
	}
	emptyStrList3, _ := types.ListValue(types.StringType, []attr.Value{})
	emptyIPRangeList3, _ := types.ListValue(FromToObjectType, []attr.Value{})
	emptyNameIDSet3, _ := types.SetValue(NameIDObjectType, []attr.Value{})
	attrs := map[string]attr.Value{
		"application":              parseNameIDListOrEmptySet(ctx, apps, "file_rule.application.application"),
		"custom_app":               parseNameIDListOrEmptySet(ctx, customs, "file_rule.application.custom_app"),
		"app_category":             parseNameIDListOrEmptySet(ctx, cats, "file_rule.application.app_category"),
		"custom_category":          parseNameIDListOrEmptySet(ctx, customCats, "file_rule.application.custom_category"),
		"sanctioned_apps_category": parseNameIDListOrEmptySet(ctx, sac, "file_rule.application.sanctioned_apps_category"),
		"domain":                   emptyStrList3,
		"fqdn":                     emptyStrList3,
		"ip":                       emptyStrList3,
		"subnet":                   emptyStrList3,
		"ip_range":                 emptyIPRangeList3,
		"global_ip_range":          emptyNameIDSet3,
	}
	o, d := types.ObjectValue(WanApplicationAttrTypes, attrs)
	diags.Append(d...)
	return o
}

func acAccessMethodValueString(
	row *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_AccessMethod,
) types.String {
	if row == nil {
		return types.StringNull()
	}
	if row.GetValue() != nil {
		return types.StringValue(string(*row.GetValue()))
	}
	if vs := row.GetValueSet(); vs != nil {
		if vs.GetID() != "" {
			return types.StringValue(vs.GetID())
		}
		return types.StringValue(vs.GetName())
	}
	return types.StringNull()
}

func acAccessMethodListFromApplicationRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_AccessMethod,
	diags *diag.Diagnostics,
) types.List {
	elemType := types.ObjectType{AttrTypes: ApplicationControlAccessMethodAttrTypes}
	if len(rows) == 0 {
		return types.ListNull(elemType)
	}
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		o, d := types.ObjectValue(ApplicationControlAccessMethodAttrTypes, map[string]attr.Value{
			"access_method": types.StringValue(row.GetAccessMethod().String()),
			"operator":      types.StringValue(row.GetOperator().String()),
			"value":         acAccessMethodValueString(row),
		})
		diags.Append(d...)
		if d.HasError() {
			continue
		}
		elems = append(elems, o)
	}
	if len(elems) == 0 {
		return types.ListNull(elemType)
	}
	lst, d := types.ListValue(elemType, elems)
	diags.Append(d...)
	if d.HasError() || lst.IsUnknown() {
		return types.ListNull(elemType)
	}
	return lst
}

func acAccessMethodValueStringData(
	row *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_AccessMethod,
) types.String {
	if row == nil {
		return types.StringNull()
	}
	if row.GetValue() != nil {
		return types.StringValue(string(*row.GetValue()))
	}
	if vs := row.GetValueSet(); vs != nil {
		if vs.GetID() != "" {
			return types.StringValue(vs.GetID())
		}
		return types.StringValue(vs.GetName())
	}
	return types.StringNull()
}

func acAccessMethodListFromDataRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_AccessMethod,
	diags *diag.Diagnostics,
) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		o, d := types.ObjectValue(ApplicationControlAccessMethodAttrTypes, map[string]attr.Value{
			"access_method": types.StringValue(row.GetAccessMethod().String()),
			"operator":      types.StringValue(row.GetOperator().String()),
			"value":         acAccessMethodValueStringData(row),
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: ApplicationControlAccessMethodAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

func acAccessMethodValueStringFile(
	row *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_AccessMethod,
) types.String {
	if row == nil {
		return types.StringNull()
	}
	if row.GetValue() != nil {
		return types.StringValue(string(*row.GetValue()))
	}
	if vs := row.GetValueSet(); vs != nil {
		if vs.GetID() != "" {
			return types.StringValue(vs.GetID())
		}
		return types.StringValue(vs.GetName())
	}
	return types.StringNull()
}

func acAccessMethodListFromFileRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_AccessMethod,
	diags *diag.Diagnostics,
) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		o, d := types.ObjectValue(ApplicationControlAccessMethodAttrTypes, map[string]attr.Value{
			"access_method": types.StringValue(row.GetAccessMethod().String()),
			"operator":      types.StringValue(row.GetOperator().String()),
			"value":         acAccessMethodValueStringFile(row),
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: ApplicationControlAccessMethodAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

func acFileAttributeListFromDataRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_FileAttribute,
	diags *diag.Diagnostics,
) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		val := types.StringNull()
		if row.GetValue() != nil {
			val = types.StringValue(string(*row.GetValue()))
		}
		o, d := types.ObjectValue(applicationControlFileAttributeAttrTypes, map[string]attr.Value{
			"file_attribute": types.StringValue(row.GetFileAttribute().String()),
			"operator":       types.StringValue(row.GetOperator().String()),
			"value":          val,
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

func acFileAttributeListFromFileRule(
	_ context.Context,
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_FileAttribute,
	diags *diag.Diagnostics,
) types.List {
	elems := make([]attr.Value, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		val := types.StringNull()
		if row.GetValue() != nil {
			val = types.StringValue(string(*row.GetValue()))
		}
		o, d := types.ObjectValue(applicationControlFileAttributeAttrTypes, map[string]attr.Value{
			"file_attribute": types.StringValue(row.GetFileAttribute().String()),
			"operator":       types.StringValue(row.GetOperator().String()),
			"value":          val,
		})
		diags.Append(d...)
		elems = append(elems, o)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: applicationControlFileAttributeAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

// acBuildActivityList builds the application_activity list from pre-extracted (id, name) pairs.
func acBuildActivityList(pairs [][2]string, diags *diag.Diagnostics) types.List {
	elems := make([]attr.Value, 0, len(pairs))
	for _, p := range pairs {
		actObj, d := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(p[0]),
			"name": types.StringValue(p[1]),
		})
		diags.Append(d...)
		outer, d2 := types.ObjectValue(applicationControlActivityAttrTypes, map[string]attr.Value{
			"activity": actObj,
		})
		diags.Append(d2...)
		elems = append(elems, outer)
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: applicationControlActivityAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

func acApplicationActivityListFromApplicationRule(
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_ApplicationActivity,
	diags *diag.Diagnostics,
) types.List {
	pairs := make([][2]string, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		pairs = append(pairs, [2]string{a.GetID(), a.GetName()})
	}
	return acBuildActivityList(pairs, diags)
}

func acApplicationActivityListFromDataRule(
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_ApplicationActivity,
	diags *diag.Diagnostics,
) types.List {
	pairs := make([][2]string, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		pairs = append(pairs, [2]string{a.GetID(), a.GetName()})
	}
	return acBuildActivityList(pairs, diags)
}

func acApplicationActivityListFromFileRule(
	rows []*cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_ApplicationActivity,
	diags *diag.Diagnostics,
) types.List {
	pairs := make([][2]string, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		a := r.GetActivity()
		pairs = append(pairs, [2]string{a.GetID(), a.GetName()})
	}
	return acBuildActivityList(pairs, diags)
}
