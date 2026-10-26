package parule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	scalars "github.com/catonetworks/cato-go-sdk/scalars"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *privAccessRuleResource) parseSource(
	ctx context.Context,
	src privateAccessRuleSource,
	diags *diag.Diagnostics,
) types.Object {
	tfSource := Source{
		Users:      idname.ParseIDNameSet(ctx, src.User, diags),
		UserGroups: idname.ParseIDNameSet(ctx, src.UsersGroup, diags),
	}
	obj, objDiags := types.ObjectValueFrom(ctx, SourceTypes, tfSource)
	diags.Append(objDiags...)
	return obj
}

func (r *privAccessRuleResource) prepareSource(
	ctx context.Context,
	src types.Object,
	diags *diag.Diagnostics,
) *cato_models.PrivateAccessPolicySourceInput {
	var userInput []*cato_models.UserRefInput
	var groupInput []*cato_models.UsersGroupRefInput

	if !utils.HasValue(src) {
		return nil
	}

	var tfSource Source
	if utils.CheckErr(diags, src.As(ctx, &tfSource, basetypes.ObjectAsOptions{})) {
		return nil
	}

	// user list
	userInput = idname.PrepareIDNameSet[cato_models.UserRefInput](ctx, tfSource.Users, diags)
	if diags.HasError() {
		return nil
	}

	// group list
	groupInput = idname.PrepareIDNameSet[cato_models.UsersGroupRefInput](ctx, tfSource.UserGroups, diags)
	if diags.HasError() {
		return nil
	}

	return &cato_models.PrivateAccessPolicySourceInput{
		User:       userInput,
		UsersGroup: groupInput,
	}
}

func (r *privAccessRuleResource) prepareSourceUpdate(
	ctx context.Context,
	src types.Object,
	diags *diag.Diagnostics,
) *cato_models.PrivateAccessPolicySourceUpdateInput {
	upd := r.prepareSource(ctx, src, diags)
	if upd == nil {
		return nil
	}
	return &cato_models.PrivateAccessPolicySourceUpdateInput{User: upd.User, UsersGroup: upd.UsersGroup}
}

func (r *privAccessRuleResource) prepareApplications(
	ctx context.Context,
	apps types.Set,
	diags *diag.Diagnostics,
) *cato_models.PrivateAccessPolicyApplicationInput {
	if !utils.HasValue(apps) {
		return nil
	}
	applicationInput := idname.PrepareIDNameSet[cato_models.PrivateApplicationRefInput](ctx, apps, diags)
	if diags.HasError() {
		return nil
	}

	return &cato_models.PrivateAccessPolicyApplicationInput{Application: applicationInput}
}

func (r *privAccessRuleResource) prepareApplicationsUpdate(
	ctx context.Context,
	apps types.Set,
	diags *diag.Diagnostics,
) *cato_models.PrivateAccessPolicyApplicationUpdateInput {
	upd := r.prepareApplications(ctx, apps, diags)
	if upd == nil {
		return nil
	}
	return &cato_models.PrivateAccessPolicyApplicationUpdateInput{Application: upd.Application}
}

func (r *privAccessRuleResource) prepareTracking(
	ctx context.Context,
	t types.Object,
	diags *diag.Diagnostics,
) *cato_models.PolicyTrackingInput {
	sdkTracking := cato_models.PolicyTrackingInput{
		Alert: &cato_models.PolicyRuleTrackingAlertInput{Enabled: false, Frequency: cato_models.PolicyRuleTrackingFrequencyEnumHourly},
		Event: &cato_models.PolicyRuleTrackingEventInput{Enabled: false},
	}

	if !utils.HasValue(t) {
		return &sdkTracking // empty object, with enabled = false
	}

	var tfTracking PolicyRuleTracking
	diags.Append(t.As(ctx, &tfTracking, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	var tfEvent PolicyRuleTrackingEvent
	diags.Append(tfTracking.Event.As(ctx, &tfEvent, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	sdkTracking.Event.Enabled = tfEvent.Enabled.ValueBool()

	var tfAlert PoliciRuleTrackingAlert
	diags.Append(tfTracking.Alert.As(ctx, &tfAlert, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	sdkTracking.Alert.Enabled = tfAlert.Enabled.ValueBool()
	sdkTracking.Alert.Frequency = cato_models.PolicyRuleTrackingFrequencyEnum(tfAlert.Frequency.ValueString())

	// Mailing lists
	sdkTracking.Alert.MailingList = idname.PrepareIDNameSet[cato_models.SubscriptionMailingListRefInput](
		ctx,
		tfAlert.MailingList,
		diags,
	)
	if diags.HasError() {
		return nil
	}

	// Subscription groups
	sdkTracking.Alert.SubscriptionGroup = idname.PrepareIDNameSet[cato_models.SubscriptionGroupRefInput](
		ctx,
		tfAlert.SubscriptionGroup,
		diags,
	)
	if diags.HasError() {
		return nil
	}

	// Webhooks
	sdkTracking.Alert.Webhook = idname.PrepareIDNameSet[cato_models.SubscriptionWebhookRefInput](
		ctx,
		tfAlert.Webhook,
		diags,
	)
	if diags.HasError() {
		return nil
	}

	return &sdkTracking
}

func (r *privAccessRuleResource) prepareTrackingUpdate(
	ctx context.Context,
	t types.Object,
	diags *diag.Diagnostics,
) *cato_models.PolicyTrackingUpdateInput {
	upd := r.prepareTracking(ctx, t, diags)
	if upd == nil {
		return nil
	}
	return &cato_models.PolicyTrackingUpdateInput{
		Alert: &cato_models.PolicyRuleTrackingAlertUpdateInput{
			Enabled:           &upd.Alert.Enabled,
			Frequency:         &upd.Alert.Frequency,
			MailingList:       upd.Alert.MailingList,
			SubscriptionGroup: upd.Alert.SubscriptionGroup,
			Webhook:           upd.Alert.Webhook,
		},
		Event: &cato_models.PolicyRuleTrackingEventUpdateInput{
			Enabled: new(upd.Event.Enabled),
		},
	}
}

func (r *privAccessRuleResource) prepareUserAttributes(
	ctx context.Context,
	uas types.Object,
	diags *diag.Diagnostics,
) *cato_models.PrivateAccessUserAttributesInput {
	// Default values
	attr := cato_models.PrivateAccessUserAttributesInput{
		RiskScore: &cato_models.RiskScoreConditionInput{
			Category: cato_models.RiskScoreCategoryAny,
			Operator: cato_models.RiskScoreOperatorGte,
		},
	}

	// User Attributes
	if !utils.HasValue(uas) {
		return &attr
	}
	var tfUserAttributes UserAttributes
	if utils.CheckErr(diags, uas.As(ctx, &tfUserAttributes, basetypes.ObjectAsOptions{})) {
		return nil
	}

	// Risk Score
	if !utils.HasValue(tfUserAttributes.RiskScore) {
		return &attr
	}
	var tfRiskScore RiskScore
	if utils.CheckErr(diags, tfUserAttributes.RiskScore.As(ctx, &tfRiskScore, basetypes.ObjectAsOptions{})) {
		return &attr
	}

	attr.RiskScore.Category = cato_models.RiskScoreCategory(tfRiskScore.Category.ValueString())
	attr.RiskScore.Operator = cato_models.RiskScoreOperator(tfRiskScore.Operator.ValueString())
	return &attr
}

func (r *privAccessRuleResource) prepareUserAttributesUpdate(
	ctx context.Context,
	uas types.Object,
	diags *diag.Diagnostics,
) *cato_models.PrivateAccessUserAttributesUpdateInput {
	upd := r.prepareUserAttributes(ctx, uas, diags)
	if upd == nil {
		return nil
	}
	return &cato_models.PrivateAccessUserAttributesUpdateInput{
		RiskScore: &cato_models.RiskScoreConditionUpdateInput{
			Category: &upd.RiskScore.Category,
			Operator: &upd.RiskScore.Operator,
		},
	}
}

func (r *privAccessRuleResource) prepareSchedule(
	ctx context.Context,
	sch types.Object,
	diags *diag.Diagnostics,
) *cato_models.PolicyScheduleInput {
	schedule := cato_models.PolicyScheduleInput{
		ActiveOn: cato_models.PolicyActiveOnEnumAlways,
	}
	if !utils.HasValue(sch) {
		return &schedule
	}

	var tfSchedule PolicySchedule
	if utils.CheckErr(diags, sch.As(ctx, &tfSchedule, basetypes.ObjectAsOptions{})) {
		return nil
	}

	schedule.ActiveOn = cato_models.PolicyActiveOnEnum(tfSchedule.ActiveOn.ValueString())

	// Custom Recurring
	if utils.HasValue(tfSchedule.CustomRecurring) {
		var tfRecuring PolicyCustomRecurring
		if utils.CheckErr(diags, tfSchedule.CustomRecurring.As(ctx, &tfRecuring, basetypes.ObjectAsOptions{})) {
			return nil
		}
		schedule.CustomRecurring = &cato_models.PolicyCustomRecurringInput{}

		// From
		schedule.CustomRecurring.From = scalars.Time(tfRecuring.From.ValueString())

		// To
		schedule.CustomRecurring.To = scalars.Time(tfRecuring.To.ValueString())

		// Days
		var days []types.String
		if utils.CheckErr(diags, tfRecuring.Days.ElementsAs(ctx, &days, false)) {
			return nil
		}
		for _, d := range days {
			if utils.HasValue(d) {
				schedule.CustomRecurring.Days = append(schedule.CustomRecurring.Days, cato_models.DayOfWeek(d.ValueString()))
			}
		}
	}

	// Custom Timeframe
	if utils.HasValue(tfSchedule.CustomTimeframe) {
		var tfTimeframe PolicyCustomTimeframe
		if utils.CheckErr(diags, tfSchedule.CustomTimeframe.As(ctx, &tfTimeframe, basetypes.ObjectAsOptions{})) {
			return nil
		}
		schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{}

		// From
		schedule.CustomTimeframe.From = tfTimeframe.From.ValueString()

		// To
		schedule.CustomTimeframe.To = tfTimeframe.To.ValueString()
	}

	return &schedule
}

func (r *privAccessRuleResource) prepareScheduleUpdate(
	ctx context.Context,
	sch types.Object,
	diags *diag.Diagnostics,
) *cato_models.PolicyScheduleUpdateInput {
	upd := r.prepareSchedule(ctx, sch, diags)
	if upd == nil {
		return nil
	}
	schedInput := cato_models.PolicyScheduleUpdateInput{ActiveOn: &upd.ActiveOn}
	if upd.CustomRecurring != nil {
		schedInput.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{
			Days: upd.CustomRecurring.Days,
			From: &upd.CustomRecurring.From,
			To:   &upd.CustomRecurring.To,
		}
	}
	if upd.CustomTimeframe != nil {
		schedInput.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{
			From: &upd.CustomTimeframe.From,
			To:   &upd.CustomTimeframe.To,
		}
	}
	return &schedInput
}

func (r *privAccessRuleResource) prepareActivePeriod(
	ctx context.Context,
	ap types.Object,
	diags *diag.Diagnostics,
) *cato_models.PolicyRuleActivePeriodInput {
	if ap.IsUnknown() || ap.IsNull() {
		return &cato_models.PolicyRuleActivePeriodInput{}
	}

	var tfPeriod PolicyRuleActivePeriod
	diags.Append(ap.As(ctx, &tfPeriod, basetypes.ObjectAsOptions{})...)

	sdkPeriod := cato_models.PolicyRuleActivePeriodInput{
		EffectiveFrom:    convert.KnownStringPointer(tfPeriod.EffectiveFrom),
		ExpiresAt:        convert.KnownStringPointer(tfPeriod.ExpiresAt),
		UseEffectiveFrom: tfPeriod.UseEffectiveFrom.ValueBool(),
		UseExpiresAt:     tfPeriod.UseExpiresAt.ValueBool(),
	}

	return &sdkPeriod
}

func (r *privAccessRuleResource) prepareActivePeriodUpdate(
	ctx context.Context,
	ap types.Object,
	diags *diag.Diagnostics,
) *cato_models.PolicyRuleActivePeriodUpdateInput {
	upd := r.prepareActivePeriod(ctx, ap, diags)
	if upd == nil {
		return nil
	}
	return &cato_models.PolicyRuleActivePeriodUpdateInput{
		EffectiveFrom:    upd.EffectiveFrom,
		ExpiresAt:        upd.ExpiresAt,
		UseEffectiveFrom: &upd.UseEffectiveFrom,
		UseExpiresAt:     &upd.UseExpiresAt,
	}
}

func (r *privAccessRuleResource) preparePlatforms(
	ctx context.Context,
	platforms types.Set,
	diags *diag.Diagnostics,
) []cato_models.OperatingSystem {
	return convert.PrepareStrings[cato_models.OperatingSystem](ctx, platforms, diags)
}

func (r *privAccessRuleResource) prepareCountries(
	ctx context.Context,
	countries types.Set,
	diags *diag.Diagnostics,
) []*cato_models.CountryRefInput {
	return idname.PrepareIDNameSet[cato_models.CountryRefInput](ctx, countries, diags)
}

func (r *privAccessRuleResource) prepareConnectionOrigins(
	ctx context.Context,
	os types.Set,
	diags *diag.Diagnostics,
) []cato_models.ConnectionOriginsEnum {
	return convert.PrepareStrings[cato_models.ConnectionOriginsEnum](ctx, os, diags)
}

func (r *privAccessRuleResource) prepareAction(action types.String) *cato_models.PrivateAccessPolicyActionInput {
	return &cato_models.PrivateAccessPolicyActionInput{Action: cato_models.PrivateAccessPolicyActionEnum(action.ValueString())}
}

func (r *privAccessRuleResource) prepareActionUpdate(action types.String) *cato_models.PrivateAccessPolicyActionUpdateInput {
	return &cato_models.PrivateAccessPolicyActionUpdateInput{Action: new(cato_models.PrivateAccessPolicyActionEnum(action.ValueString()))}
}

func (r *privAccessRuleResource) prepareDevice(
	ctx context.Context,
	devs types.Set,
	diags *diag.Diagnostics,
) []*cato_models.DeviceProfileRefInput {
	return idname.PrepareIDNameSet[cato_models.DeviceProfileRefInput](ctx, devs, diags)
}

func (r *privAccessRuleResource) parseTracking(
	ctx context.Context,
	tr privateAccessRuleTracking,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics

	// Prepare Tracking.Event object
	tfEvent := PolicyRuleTrackingEvent{
		Enabled: types.BoolValue(tr.Event.Enabled),
	}
	eventObj, objDiags := types.ObjectValueFrom(ctx, PolicyRuleTrackingEventTypes, tfEvent)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleTrackingTypes)
	}

	// Prepare Tracking.Alert object
	mailingList := idname.ParseIDNameSet(ctx, tr.Alert.MailingList, diags)
	subscriptionGroup := idname.ParseIDNameSet(ctx, tr.Alert.SubscriptionGroup, diags)
	webHook := idname.ParseIDNameSet(ctx, tr.Alert.Webhook, diags)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleTrackingTypes)
	}
	tfAlert := PoliciRuleTrackingAlert{
		Enabled:           types.BoolValue(tr.Alert.Enabled),
		Frequency:         types.StringValue(string(tr.Alert.Frequency)),
		MailingList:       mailingList,
		SubscriptionGroup: subscriptionGroup,
		Webhook:           webHook,
	}
	alertObj, objDiags := types.ObjectValueFrom(ctx, PolicyRuleTrackingAlertTypes, tfAlert)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleTrackingTypes)
	}

	// Prepare Tracking object
	tfTracking := PolicyRuleTracking{
		Event: eventObj,
		Alert: alertObj,
	}
	trackingObj, objDiags := types.ObjectValueFrom(ctx, PolicyRuleTrackingTypes, tfTracking)
	diags.Append(objDiags...)

	return trackingObj
}

func (r *privAccessRuleResource) parseUserAttributes(
	ctx context.Context,
	ua privateAccessRuleUserAttributes,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics

	// Prepare UserAttributes.RiskScore object
	tfRiskScore := RiskScore{
		Category: types.StringValue(string(ua.RiskScore.Category)),
		Operator: types.StringValue(string(ua.RiskScore.Operator)),
	}
	riskScoreObj, objDiags := types.ObjectValueFrom(ctx, RiskScoreTypes, tfRiskScore)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(UserAttributesTypes)
	}

	// Prepare UserAttributes object
	tfUserAttributes := UserAttributes{
		RiskScore: riskScoreObj,
	}
	userAttrObj, objDiags := types.ObjectValueFrom(ctx, UserAttributesTypes, tfUserAttributes)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleTrackingTypes)
	}

	return userAttrObj
}

func (r *privAccessRuleResource) parsePolicySchedule(
	ctx context.Context,
	sch privateAccessRuleSchedule,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics

	// Prepare PolicySchedule.PolicyCustomRecurring object
	recurringObj := types.ObjectNull(PolicyCustomRecurringTypes)
	if sch.CustomRecurring != nil {
		tfRecurring := PolicyCustomRecurring{
			Days: convert.StringSet(ctx, sch.CustomRecurring.Days, diags),
			From: types.StringValue(string(sch.CustomRecurring.From)),
			To:   types.StringValue(string(sch.CustomRecurring.To)),
		}
		recurringObj, objDiags = types.ObjectValueFrom(ctx, PolicyCustomRecurringTypes, tfRecurring)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ObjectNull(PolicyScheduleTypes)
		}
	}

	// Prepare PolicySchedule.PolicyCustomTimeframe object
	timeframeObj := types.ObjectNull(PolicyCustomTimeframeTypes)
	if sch.CustomTimeframe != nil {
		tfTimeframe := PolicyCustomTimeframe{
			From: types.StringValue(utils.NormalizeDateTime(sch.CustomTimeframe.From)),
			To:   types.StringValue(utils.NormalizeDateTime(sch.CustomTimeframe.To)),
		}
		timeframeObj, objDiags = types.ObjectValueFrom(ctx, PolicyCustomTimeframeTypes, tfTimeframe)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ObjectNull(PolicyScheduleTypes)
		}
	}

	// Prepare PolicySchedule
	tfSchedule := PolicySchedule{
		ActiveOn:        types.StringValue(string(sch.ActiveOn)),
		CustomRecurring: recurringObj,
		CustomTimeframe: timeframeObj,
	}
	scheduleObj, objDiags := types.ObjectValueFrom(ctx, PolicyScheduleTypes, tfSchedule)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyScheduleTypes)
	}

	return scheduleObj
}

func (r *privAccessRuleResource) parsePolicyActivePeriod(
	ctx context.Context,
	ap privateAccessRuleActivePeriod,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics

	// Prepare Active Period
	tfActivePeriod := PolicyRuleActivePeriod{
		EffectiveFrom:    types.StringPointerValue(utils.NormalizeDateTimePtr(ap.EffectiveFrom)),
		ExpiresAt:        types.StringPointerValue(utils.NormalizeDateTimePtr(ap.ExpiresAt)),
		UseEffectiveFrom: types.BoolValue(ap.UseEffectiveFrom),
		UseExpiresAt:     types.BoolValue(ap.UseExpiresAt),
	}
	periodObj, objDiags := types.ObjectValueFrom(ctx, PolicyRuleActivePeriodTypes, tfActivePeriod)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleActivePeriodTypes)
	}

	return periodObj
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	r *privAccessRuleResource,
	plan RuleModel,
	diagnostics *diag.Diagnostics,
) (resultRuleName string, resultInput cato_models.PrivateAccessAddRuleInput) {
	resultRuleName = plan.Name.ValueString()
	resultInput = cato_models.PrivateAccessAddRuleInput{
		Rule: &cato_models.PrivateAccessAddRuleDataInput{
			Action:                r.prepareAction(plan.Action),
			ActivePeriod:          r.prepareActivePeriod(ctx, plan.ActivePeriod, diagnostics),
			Applications:          r.prepareApplications(ctx, plan.Applications, diagnostics),
			ConnectionsOriginList: r.prepareConnectionOrigins(ctx, plan.ConnectionOrigins, diagnostics),
			Country:               r.prepareCountries(ctx, plan.Countries, diagnostics),
			Description:           plan.Description.ValueString(),
			Device:                r.prepareDevice(ctx, plan.Devices, diagnostics),
			Enabled:               plan.Enabled.ValueBool(),
			Name:                  resultRuleName,
			Platform:              r.preparePlatforms(ctx, plan.Platforms, diagnostics),
			Schedule:              r.prepareSchedule(ctx, plan.Schedule, diagnostics),
			Source:                r.prepareSource(ctx, plan.Source, diagnostics),
			Tracking:              r.prepareTracking(ctx, plan.Tracking, diagnostics),
			UserAttributes:        r.prepareUserAttributes(ctx, plan.UserAttributes, diagnostics),
		},
		At: &cato_models.PolicyRulePositionInput{
			Position: new(cato_models.PolicyRulePositionEnumLastInPolicy),
		},
	}
	return resultRuleName, resultInput
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state RuleModel) (resultInput cato_models.PrivateAccessRemoveRuleInput) {
	resultInput = cato_models.PrivateAccessRemoveRuleInput{ID: state.ID.ValueString()}
	return resultInput
}

// convertHydratePrivAccessRuleStateValues converts between the Terraform model and SDK values.
func convertHydratePrivAccessRuleStateValues(
	ctx context.Context,
	r *privAccessRuleResource,
	originalDiags *diag.Diagnostics,
	priorState **RuleModel,
	polRule *cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules,
) {
	diags := *originalDiags
	defer func() { *originalDiags = diags }()

	state := *priorState
	defer func() { *priorState = state }()
	apiRule := polRule.Rule
	state = &RuleModel{
		Action:            types.StringValue(string(apiRule.Action.Action)),
		ActivePeriod:      r.parsePolicyActivePeriod(ctx, apiRule.ActivePeriod, &diags),
		Applications:      idname.ParseIDNameSet(ctx, apiRule.Applications.Application, &diags),
		ConnectionOrigins: convert.StringSet(ctx, apiRule.ConnectionsOriginList, &diags),
		Countries:         idname.ParseIDNameSet(ctx, apiRule.Country, &diags),
		Description:       types.StringValue(apiRule.Description),
		Devices:           idname.ParseIDNameSet(ctx, apiRule.Device, &diags),
		Enabled:           types.BoolValue(apiRule.Enabled),
		ID:                types.StringValue(apiRule.ID),
		Name:              types.StringValue(apiRule.Name),
		Platforms:         convert.StringSet(ctx, apiRule.Platform, &diags),
		Schedule:          r.parsePolicySchedule(ctx, apiRule.Schedule, &diags),
		Source:            r.parseSource(ctx, apiRule.Source, &diags),
		Tracking:          r.parseTracking(ctx, apiRule.Tracking, &diags),
		UserAttributes:    r.parseUserAttributes(ctx, apiRule.UserAttributes, &diags),
	}
}
