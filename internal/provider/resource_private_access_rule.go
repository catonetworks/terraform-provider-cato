package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &privAccessRuleResource{}
	_ resource.ResourceWithConfigure   = &privAccessRuleResource{}
	_ resource.ResourceWithImportState = &privAccessRuleResource{}

	ErrPrivateAcccessRuleNotFound = errors.New("private access rule not found")
)

func NewPrivAccessRuleResource() resource.Resource {
	return &privAccessRuleResource{}
}

type privAccessRuleResource struct {
	client *catoClientData
}

func (r *privAccessRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_access_rule"
}

func schemaNameID(prefix string) map[string]schema.Attribute {
	if prefix != "" {
		prefix += " "
	}
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: prefix + "name",
			Optional:    true,
			Computed:    true,
		},
		"id": schema.StringAttribute{
			Description: prefix + "ID",
			Optional:    true,
			Computed:    true,
		},
	}
}

func (r *privAccessRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_access_rule` resource contains the configuration parameters for private access policy rule in the Cato platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Rule ID",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Rule name",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Rule description",
				Optional:    true,
			},
			"index": schema.Int64Attribute{
				Description: "Rule index",
				Required:    true,
			},
			// "section": r.schemaSection(), -- not available in the 1st phase
			"enabled": schema.BoolAttribute{
				Description: "TRUE = Rule is enabled FALSE = Rule is disabled",
				Required:    true,
			},
			"source": r.schemaSource(),
			"platforms": schema.ListAttribute{
				Description: "Platforms, operating systems",
				Optional:    true,
				ElementType: types.StringType,
			},
			"countries": schema.ListNestedAttribute{
				Description:  "Country name or id",
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("Country")}},
			"applications": schema.ListNestedAttribute{
				Description:  "Application name or id",
				Required:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("Application")},
			},
			"connection_origin": schema.ListAttribute{
				Description: "Origin of the connection",
				Optional:    true,
				ElementType: types.StringType,
			},
			"action": schema.StringAttribute{
				Description: "ALLOW or BLOCK",
				Required:    true,
				Validators:  []validator.String{privAccPolicyActionValidator{}},
			},
			"tracking": r.schemaTracking(),
			"device": schema.ListNestedAttribute{
				Description:  "Device group name or id",
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("Device group")},
			},
			"user_attributes": r.schemaUserAttributes(),
			"schedule":        r.schemaSchedule(),
			"active_period":   r.schemaActivePeriod(),
		},
	}
}

func (r *privAccessRuleResource) schemaSection() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Settings for a rule section",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Section ID",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Section name",
				Required:    true,
			},
			"subrule_id": schema.StringAttribute{
				Description: "Subrule ID",
				Optional:    true,
			},
		},
	}
}

func (r *privAccessRuleResource) schemaTracking() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Rule tracking",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"event": schema.SingleNestedAttribute{
				Description: "Event settings",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Event tracking enabled",
						Required:    true,
					},
				},
			},
			"alert": schema.SingleNestedAttribute{
				Description: "Alert settings",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "TRUE – send alerts when the rule is matched, FALSE – don’t send alerts when the rule is matched",
						Required:    true,
					},
					"frequency": schema.StringAttribute{
						Description: "Frequency of an alert event for a rule",
						Required:    true,
					},

					"subscription_group": schema.ListNestedAttribute{
						Description:  "Subscription group name or id",
						Optional:     true,
						NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("Subscription group")},
					},
					"webhook": schema.ListNestedAttribute{
						Description:  "Webhook name or id",
						Optional:     true,
						NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("Webhook")},
					},
					"mailing_list": schema.ListNestedAttribute{
						Description:  "Mailing list name or id",
						Optional:     true,
						NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("Mailing list")},
					},
				},
			},
		},
	}
}

func (r *privAccessRuleResource) schemaUserAttributes() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "User attributes",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"risk_score": schema.SingleNestedAttribute{
				Description: "User's risk score settings",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"category": schema.StringAttribute{
						Description: "Risk score category",
						Required:    true,
					},
					"operator": schema.StringAttribute{
						Description: "Risk score operator",
						Required:    true,
					},
				},
			},
		},
	}
}

func (r *privAccessRuleResource) schemaSchedule() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Schedule",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"active_on": schema.StringAttribute{
				Description: "Type of a time range when a rule is active",
				Required:    true,
			},
			"custom_recurring": schema.SingleNestedAttribute{
				Description: "Custom recurring time range that a rule is active",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"days": schema.ListAttribute{
						Description: "Days of the week",
						Required:    true,
						ElementType: types.StringType,
					},
					"from": schema.StringAttribute{
						Description: "From",
						Optional:    true,
					},
					"to": schema.StringAttribute{
						Description: "To",
						Optional:    true,
					},
				},
			},
			"custom_timeframe": schema.SingleNestedAttribute{
				Description: "Custom one-time time range that a rule is active",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"from": schema.StringAttribute{
						Description: "From",
						Required:    true,
					},
					"to": schema.StringAttribute{
						Description: "To",
						Required:    true,
					},
				},
			},
		},
	}
}

func (r *privAccessRuleResource) schemaActivePeriod() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Time period during which the rule is active",
		Optional:    true,
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"effective_from": schema.StringAttribute{
				Description: "Effective from",
				Optional:    true,
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "Expires at",
				Optional:    true,
				Computed:    true,
			},
			"use_effective_from": schema.BoolAttribute{
				Description: "Use effective from",
				Optional:    true,
				Computed:    true,
			},
			"use_expires_at": schema.BoolAttribute{
				Description: "Use expires at",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *privAccessRuleResource) schemaSource() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Source",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Description:  "Users",
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("User")},
			},
			"user_groups": schema.ListNestedAttribute{
				Description:  "User groups",
				Optional:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: schemaNameID("Group")},
			},
		},
		Validators: []validator.Object{privAccPolicySourceValidator{}},
	}
}

func (r *privAccessRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *privAccessRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *privAccessRuleResource) prepareSource(src *Source, diags *diag.Diagnostics) *cato_models.PrivateAccessPolicySourceInput {
	var userInput []*cato_models.UserRefInput
	var groupInput []*cato_models.UsersGroupRefInput

	// user list
	for _, usr := range src.Users {
		refBy, refInput, _ := prepareIdName(usr.ID, usr.Name, diags, "source.users")
		if diags.HasError() {
			return nil
		}
		userInput = append(userInput, &cato_models.UserRefInput{By: refBy, Input: refInput})
	}

	// group list
	for _, grp := range src.UserGroups {
		refBy, refInput, _ := prepareIdName(grp.ID, grp.Name, diags, "source.groups")
		if diags.HasError() {
			return nil
		}
		groupInput = append(groupInput, &cato_models.UsersGroupRefInput{By: refBy, Input: refInput})
	}
	return &cato_models.PrivateAccessPolicySourceInput{
		User:       userInput,
		UsersGroup: groupInput,
	}
}

func (r *privAccessRuleResource) prepareApplications(apps []IdNameRefModel, diags *diag.Diagnostics) *cato_models.PrivateAccessPolicyApplicationInput {
	var applicationInput []*cato_models.PrivateApplicationRefInput

	// app list
	for _, app := range apps {
		refBy, refInput, _ := prepareIdName(app.ID, app.Name, diags, "applications")
		if diags.HasError() {
			return nil
		}
		applicationInput = append(applicationInput, &cato_models.PrivateApplicationRefInput{By: refBy, Input: refInput})
	}

	return &cato_models.PrivateAccessPolicyApplicationInput{
		Application: applicationInput,
	}
}

func (r *privAccessRuleResource) prepareTracking(t *Tracking, diags *diag.Diagnostics) *cato_models.PolicyTrackingInput {
	tracking := cato_models.PolicyTrackingInput{
		Alert: &cato_models.PolicyRuleTrackingAlertInput{Enabled: false, Frequency: cato_models.PolicyRuleTrackingFrequencyEnumHourly},
		Event: &cato_models.PolicyRuleTrackingEventInput{Enabled: false},
	}

	if t == nil {
		return &tracking // empty object, with enabled = false
	}

	tracking.Event.Enabled = t.Event.Enabled.ValueBool()

	tracking.Alert.Enabled = t.Alert.Enabled.ValueBool()
	tracking.Alert.Frequency = cato_models.PolicyRuleTrackingFrequencyEnum(t.Alert.Frequency.ValueString())

	// Mailing lists
	for _, ml := range t.Alert.MailingList {
		refBy, refInput, _ := prepareIdName(ml.ID, ml.Name, diags, "tracking.alert.mailing_list")
		if diags.HasError() {
			return nil
		}
		tracking.Alert.MailingList = append(tracking.Alert.MailingList, &cato_models.SubscriptionMailingListRefInput{By: refBy, Input: refInput})
	}

	// Subscription groups
	for _, sg := range t.Alert.SubscriptionGroup {
		refBy, refInput, _ := prepareIdName(sg.ID, sg.Name, diags, "tracking.alert.subscription_group")
		if diags.HasError() {
			return nil
		}
		tracking.Alert.SubscriptionGroup = append(tracking.Alert.SubscriptionGroup, &cato_models.SubscriptionGroupRefInput{By: refBy, Input: refInput})
	}

	// Webhooks
	for _, wh := range t.Alert.Webhook {
		refBy, refInput, _ := prepareIdName(wh.ID, wh.Name, diags, "tracking.alert.webhook")
		if diags.HasError() {
			return nil
		}
		tracking.Alert.Webhook = append(tracking.Alert.Webhook, &cato_models.SubscriptionWebhookRefInput{By: refBy, Input: refInput})
	}

	return &tracking
}

func (r *privAccessRuleResource) prepareUserAttributes(ua *UserAttributes) *cato_models.UserAttributesInput {
	attr := cato_models.UserAttributesInput{
		RiskScore: &cato_models.RiskScoreConditionInput{
			Category: cato_models.RiskScoreCategoryAny,
			Operator: cato_models.RiskScoreOperatorGte,
		},
	}

	if ua == nil {
		return &attr
	}

	attr.RiskScore.Category = cato_models.RiskScoreCategory(ua.RiskScore.Category.ValueString())
	attr.RiskScore.Operator = cato_models.RiskScoreOperator(ua.RiskScore.Operator.ValueString())
	return &attr
}

func (r *privAccessRuleResource) prepareSchedule(sch *PolicySchedule, diags *diag.Diagnostics) *cato_models.PolicyScheduleInput {
	schedule := cato_models.PolicyScheduleInput{
		ActiveOn: cato_models.PolicyActiveOnEnumAlways,
	}
	if sch == nil {
		return &schedule
	}
	if !sch.ActiveOn.IsUnknown() {
		schedule.ActiveOn = cato_models.PolicyActiveOnEnum(sch.ActiveOn.ValueString())
	}

	// Custom Recurring
	rec := sch.CustomRecurring
	var recInput cato_models.PolicyCustomRecurringInput
	for _, d := range rec.Days {
		if !d.IsUnknown() {
			recInput.Days = append(recInput.Days, cato_models.DayOfWeek(d.ValueString()))
		}
	}
	if !rec.From.IsUnknown() && rec.From.ValueString() != "" {
		recInput.From = scalars.Time(rec.From.ValueString())
	}
	if !rec.To.IsUnknown() && rec.To.ValueString() != "" {
		recInput.To = scalars.Time(rec.To.ValueString())
	}
	if (len(recInput.Days) > 0) || (recInput.From != "") || (recInput.To != "") {
		schedule.CustomRecurring = &recInput
	}

	// Custom Timeframe
	tf := sch.CustomTimeframe
	var tfInput cato_models.PolicyCustomTimeframeInput
	if !tf.From.IsUnknown() && tf.From.ValueString() != "" {
		tfInput.From = tf.From.ValueString()
	}
	if !tf.To.IsUnknown() && tf.To.ValueString() != "" {
		tfInput.To = tf.To.ValueString()
	}
	if (tfInput.From != "") || (tfInput.To != "") {
		schedule.CustomTimeframe = &tfInput
	}

	return &schedule
}

func (r *privAccessRuleResource) prepareActivePeriod(ap types.Object, diags *diag.Diagnostics) *cato_models.PolicyRuleActivePeriodInput {
	activePeriod := cato_models.PolicyRuleActivePeriodInput{}

	attrs := ap.Attributes()
	fmt.Printf("XXX %v\n", attrs)
	attrTypes := ap.AttributeTypes(context.Background())
	fmt.Printf("XXX %v\n", attrTypes)

	if ap.IsUnknown() || ap.IsNull() {
		return &activePeriod
	}

	// if hasValue(ap.EffectiveFrom) {
	// 	activePeriod.EffectiveFrom = ap.EffectiveFrom.ValueStringPointer()
	// }
	// if hasValue(ap.ExpiresAt) {
	// 	activePeriod.ExpiresAt = ap.ExpiresAt.ValueStringPointer()
	// }
	// if hasValue(ap.UseEffectiveFrom) {
	// 	activePeriod.UseEffectiveFrom = ap.UseEffectiveFrom.ValueBool()
	// }
	// if hasValue(ap.UseExpiresAt) {
	// 	activePeriod.UseExpiresAt = ap.UseExpiresAt.ValueBool()
	// }

	return &activePeriod
}

func (r *privAccessRuleResource) preparePlatforms(pl []types.String, diags *diag.Diagnostics) []cato_models.OperatingSystem {
	platforms := make([]cato_models.OperatingSystem, 0, len(pl))
	for _, p := range pl {
		if !p.IsUnknown() {
			platforms = append(platforms, cato_models.OperatingSystem(p.ValueString()))
		}
	}
	return platforms
}

func (r *privAccessRuleResource) prepareCountries(countyRefs []IdNameRefModel, diags *diag.Diagnostics) []*cato_models.CountryRefInput {
	countries := make([]*cato_models.CountryRefInput, 0, len(countyRefs))
	for _, cRef := range countyRefs {

		refBy, refInput, _ := prepareIdName(cRef.ID, cRef.Name, diags, "rule.country")
		if diags.HasError() {
			return nil
		}
		countries = append(countries, &cato_models.CountryRefInput{By: refBy, Input: refInput})
	}
	return countries
}

func (r *privAccessRuleResource) prepareConnectionOrigins(os []types.String, diags *diag.Diagnostics) []cato_models.PrivateAccessPolicyOriginEnum {
	origins := make([]cato_models.PrivateAccessPolicyOriginEnum, 0, len(os))
	for _, origin := range os {
		if !origin.IsUnknown() {
			origins = append(origins, cato_models.PrivateAccessPolicyOriginEnum(origin.ValueString()))
		}
	}
	return origins
}

func (r *privAccessRuleResource) prepareDevice(devs []IdNameRefModel, diags *diag.Diagnostics) []*cato_models.DeviceProfileRefInput {
	devices := make([]*cato_models.DeviceProfileRefInput, 0, len(devs))
	for _, dRef := range devs {
		refBy, refInput, _ := prepareIdName(dRef.ID, dRef.Name, diags, "device")
		if diags.HasError() {
			return nil
		}
		devices = append(devices, &cato_models.DeviceProfileRefInput{By: refBy, Input: refInput})
	}
	return devices
}

func (r *privAccessRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PrivateAccessRuleModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// attrs := plan.Attributes()
	// fmt.Printf("XXX %v\n", attrs)
	// attrTypes := plan.AttributeTypes(context.Background())
	// fmt.Printf("XXX %v\n", attrTypes)

	ruleName := plan.Name.ValueString()
	input := cato_models.PrivateAccessAddRuleInput{
		Rule: &cato_models.PrivateAccessAddRuleDataInput{
			Enabled:          plan.Enabled.ValueBool(),
			Name:             ruleName,
			Description:      plan.Description.ValueString(),
			Source:           r.prepareSource(plan.Source, &resp.Diagnostics),
			Platform:         r.preparePlatforms(plan.Platforms, &resp.Diagnostics),
			Country:          r.prepareCountries(plan.Countries, &resp.Diagnostics),
			Applications:     r.prepareApplications(plan.Applications, &resp.Diagnostics),
			ConnectionOrigin: r.prepareConnectionOrigins(plan.ConnectionOrigin, &resp.Diagnostics),
			Action:           &cato_models.PrivateAccessPolicyActionInput{Action: cato_models.PrivateAccessPolicyActionEnum(plan.Action.ValueString())},
			Tracking:         r.prepareTracking(plan.Tracking, &resp.Diagnostics),
			Device:           r.prepareDevice(plan.Device, &resp.Diagnostics),
			UserAttributes:   r.prepareUserAttributes(plan.UserAttributes),
			Schedule:         r.prepareSchedule(plan.Schedule, &resp.Diagnostics),
			ActivePeriod:     r.prepareActivePeriod(plan.ActivePeriod, &resp.Diagnostics),
		},
		At: &cato_models.PolicyRulePositionInput{
			Position: ptr(cato_models.PolicyRulePositionEnumLastInPolicy),
		},
	}
	if resp.Diagnostics.HasError() {
		return
	}
	// Call Cato API to create a new rule
	tflog.Debug(ctx, "PolicyPrivateAccessAddRule", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	result, err := r.client.catov2.PolicyPrivateAccessAddRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessAddRule", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	errMsg := fmt.Sprintf("failed to add private access rule '%s'", ruleName)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	respRule := result.GetPolicy().GetPrivateAccess().GetAddRule()
	if respRule.Status != cato_models.PolicyMutationStatusSuccess {
		resp.Diagnostics.AddError(errMsg, "returned status: "+string(respRule.Status))
		for _, e := range respRule.Errors {
			resp.Diagnostics.AddError(errMsg, *e.ErrorMessage)
		}
		return
	}
	ruleID := result.GetPolicy().GetPrivateAccess().GetAddRule().GetRule().GetRule().GetID()
	if ruleID == "" {
		resp.Diagnostics.AddError("Cato API PolicyPrivateAccessAddRule error", "rule ID is empty")
		return
	}

	// Set the ID from the response
	plan.ID = types.StringValue(ruleID)

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydratePrivAccessRuleState(ctx, ruleID, plan)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateAccessRule state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *privAccessRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PrivateAccessRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, hydrateErr := r.hydratePrivAccessRuleState(ctx, state.ID.ValueString(), state)
	if hydrateErr != nil {
		if errors.Is(hydrateErr, ErrPrivateAcccessRuleNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Private access rule %s not found, resource removed", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error hydrating group state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// hydratePrivAccessRuleState fetches the current state of a privAccessRule from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessRuleResource) hydratePrivAccessRuleState(ctx context.Context, ruleID string, plan PrivateAccessRuleModel) (*PrivateAccessRuleModel, error) {
	// Call Cato API to get the policy
	result, err := r.client.catov2.PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "PolicyReadPrivateAccessPolicy", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		return nil, err
	}

	var state *PrivateAccessRuleModel

	// Map API response to PrivAccessPolicyModel
	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
	for _, polRule := range policy.Rules {
		if polRule.Rule.ID != ruleID {
			continue
		}
		apiRule := polRule.Rule
		state = &PrivateAccessRuleModel{
			ID:               types.StringValue(apiRule.ID),
			Name:             types.StringValue(apiRule.Name),
			Description:      types.StringValue(apiRule.Description),
			Index:            types.Int64Value(apiRule.Index),
			Enabled:          types.BoolValue(apiRule.Enabled),
			Source:           &Source{Users: parseIDRef(apiRule.Source.User), UserGroups: parseIDRef(apiRule.Source.UsersGroup)},
			Platforms:        parseStringList(apiRule.Platform),
			Countries:        parseIDRef(apiRule.Country),
			Applications:     parseIDRef(apiRule.Applications.Application),
			ConnectionOrigin: parseStringList(apiRule.ConnectionOrigin),
			Action:           types.StringValue(string(apiRule.Action.Action)),
			Tracking:         r.parseTracking(apiRule.Tracking),
			Device:           parseIDRef(apiRule.Device),
			UserAttributes: &UserAttributes{RiskScore: RiskScore{
				Category: types.StringValue(string(apiRule.UserAttributes.RiskScore.Category)),
				Operator: types.StringValue(string(apiRule.UserAttributes.RiskScore.Operator)),
			}},
			Schedule:     r.parsePolicySchedule(apiRule.Schedule),
			ActivePeriod: r.parsePolicyActivePeriod(plan.ActivePeriod, apiRule.ActivePeriod),
		}
		break
	}

	if state == nil {
		return nil, ErrPrivateAcccessRuleNotFound
	}

	return state, nil
}

func (r *privAccessRuleResource) parseTracking(t cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Tracking) *Tracking {
	alert := &t.Alert
	out := Tracking{
		Event: PolicyRuleTrackingEvent{Enabled: types.BoolValue(t.Event.Enabled)},
		Alert: PoliciRuleTrackingAlert{
			Enabled:           types.BoolValue(alert.Enabled),
			Frequency:         types.StringValue(string(alert.Frequency)),
			MailingList:       parseIDRef(alert.MailingList),
			SubscriptionGroup: parseIDRef(alert.MailingList),
			Webhook:           parseIDRef(alert.Webhook),
		},
	}
	return &out
}
func (r *privAccessRuleResource) parsePolicySchedule(s cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Schedule) *PolicySchedule {
	p := PolicySchedule{
		ActiveOn: types.StringValue(string(s.ActiveOn)),
	}
	if s.CustomRecurring != nil {
		cr := s.CustomRecurring
		p.CustomRecurring = PolicyCustomRecurring{
			Days: parseStringList(cr.Days),
			From: types.StringValue(string(cr.From)),
			To:   types.StringValue(string(cr.To)),
		}
	}
	if s.CustomTimeframe != nil {
		tf := s.CustomTimeframe
		p.CustomTimeframe = PolicyCustomTimeframe{
			From: types.StringValue(string(tf.From)),
			To:   types.StringValue(string(tf.To)),
		}
	}
	return &p
}

func (r *privAccessRuleResource) parsePolicyActivePeriod(obj types.Object, ap cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_ActivePeriod) types.Object {
	// p := PolicyRuleActivePeriod{
	// 	EffectiveFrom:    types.StringPointerValue(ap.EffectiveFrom),
	// 	ExpiresAt:        types.StringPointerValue(ap.ExpiresAt),
	// 	UseEffectiveFrom: types.BoolValue(ap.UseEffectiveFrom),
	// 	UseExpiresAt:     types.BoolValue(ap.UseExpiresAt),
	// }
	attrT := obj.AttributeTypes(context.Background())

	attrTypes := map[string]attr.Type{
		"effective_from":     types.StringType,
		"expires_at":         types.StringType,
		"use_effective_from": types.BoolType,
		"use_expires_at":     types.BoolType,
	}
	_ = attrT

	obj, diags := types.ObjectValue(
		attrTypes,
		map[string]attr.Value{
			"effective_from":     types.StringPointerValue(ap.EffectiveFrom),
			"expires_at":         types.StringPointerValue(ap.ExpiresAt),
			"use_effective_from": types.BoolValue(ap.UseEffectiveFrom),
			"use_expires_at":     types.BoolValue(ap.UseExpiresAt),
		},
	)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	_ = diags // TODO: implement
	return obj
}

type privAccPolicySourceValidator struct{}

// ValidateObject for the "source" ensures that there is either a user or a group specified.
func (v privAccPolicySourceValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	addError := func(msg ...string) {
		if len(msg) == 0 {
			msg = []string{"at least one user or group must be specified"}
		}
		resp.Diagnostics.AddError("Field validation error", "invalid private_acces_policy source: "+msg[0])
	}
	checkError := func(e error) bool {
		if e == nil {
			return false
		}
		addError(e.Error())
		return true
	}

	source := req.ConfigValue.Attributes()
	if source == nil {
		addError()
		return
	}
	for _, srcKind := range []string{"users", "user_groups"} {
		if attrValue := source[srcKind]; attrValue != nil {
			var items []tftypes.Value
			tfvalue, err := attrValue.ToTerraformValue(context.Background())
			if checkError(err) {
				return
			}
			if checkError(tfvalue.As(&items)) {
				return
			}
			if len(items) > 0 {
				return // Good, users or groups are specified
			}
		}
	}
	addError() // No users or groups specified
}

func (v privAccPolicySourceValidator) Description(ctx context.Context) string {
	return "PrivatAccessPolicy source must specify at least one user or group"
}
func (v privAccPolicySourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

type privAccPolicyActionValidator struct{}

func (v privAccPolicyActionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := strings.Trim(req.ConfigValue.String(), `"`)
	action := cato_models.PrivateAccessPolicyActionEnum(value)
	if !action.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid action (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllPrivateAccessPolicyActionEnum))
		return
	}
}
func (v privAccPolicyActionValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("PrivatAccessPolicy action must be one of: %v", cato_models.AllPrivateAccessPolicyActionEnum)
}
func (v privAccPolicyActionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
