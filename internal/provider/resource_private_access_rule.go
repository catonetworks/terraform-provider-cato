package provider

import (
	"context"
	"errors"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &privAccessRuleResource{}
	_ resource.ResourceWithConfigure   = &privAccessRuleResource{}
	_ resource.ResourceWithImportState = &privAccessRuleResource{}

	ErrPrivateAcccessRuleNotFound = errors.New("private access rule not found")
	ErrAPIResponseParse           = errors.New("failed to parse API response")
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

func (r *privAccessRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_access_rule` resource contains the configuration parameters for private access policy rule in the Cato platform.",
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Description: "ALLOW or BLOCK",
				Required:    true,
				Validators:  []validator.String{validators.PrivAccPolicyActionValidator{}},
			},
			"active_period": r.schemaActivePeriod(),
			"applications": schema.ListNestedAttribute{
				Description:  "Application name or id",
				Required:     true,
				NestedObject: schema.NestedAttributeObject{Attributes: parse.SchemaNameID("Application")},
			},
			"connection_origins": schema.ListAttribute{
				Description:   "Origin of the connection",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				Validators:    []validator.List{validators.PrivAccPolicyConnOriginValidator{}},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"countries": schema.ListNestedAttribute{
				Description:   "List of countries",
				Optional:      true,
				Computed:      true,
				NestedObject:  schema.NestedAttributeObject{Attributes: parse.SchemaNameID("Country")},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Description: "Rule description",
				Optional:    true,
			},
			"devices": schema.ListNestedAttribute{
				Description:   "List of devices",
				Optional:      true,
				Computed:      true,
				NestedObject:  schema.NestedAttributeObject{Attributes: parse.SchemaNameID("Device")},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Description: "TRUE = Rule is enabled FALSE = Rule is disabled",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description:   "Rule ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Rule name",
				Required:    true,
			},
			"platforms": schema.ListAttribute{
				Description:   "Platforms, operating systems",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
				Validators:    []validator.List{validators.PlatformValidator{}},
			},
			"schedule":        r.schemaSchedule(),
			"source":          r.schemaSource(),
			"tracking":        r.schemaTracking(),
			"user_attributes": r.schemaUserAttributes(),
			// index is only used in bulk operation
			// "section": r.schemaSection(), -- not available in the 1st phase
		},
	}
}

func (r *privAccessRuleResource) schemaTracking() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:   "Rule tracking",
		Optional:      true,
		Computed:      true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
		Attributes: map[string]schema.Attribute{
			"alert": schema.SingleNestedAttribute{
				Description:   "Alert settings",
				Required:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "TRUE – send alerts when the rule is matched, FALSE – don’t send alerts when the rule is matched",
						Required:    true,
					},
					"frequency": schema.StringAttribute{
						Description: "Frequency of an alert event for a rule",
						Required:    true,
						Validators:  []validator.String{validators.PolicyTrackingFrequency{}},
					},
					"mailing_list": schema.ListNestedAttribute{
						Description:  "Mailing list name or id",
						Optional:     true,
						NestedObject: schema.NestedAttributeObject{Attributes: parse.SchemaNameID("Mailing list")},
					},
					"subscription_group": schema.ListNestedAttribute{
						Description:  "Subscription group name or id",
						Optional:     true,
						NestedObject: schema.NestedAttributeObject{Attributes: parse.SchemaNameID("Subscription group")},
					},
					"webhook": schema.ListNestedAttribute{
						Description:  "Webhook name or id",
						Optional:     true,
						NestedObject: schema.NestedAttributeObject{Attributes: parse.SchemaNameID("Webhook")},
					},
				},
			},
			"event": schema.SingleNestedAttribute{
				Description: "Event settings",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description:   "Event tracking enabled",
						Required:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
				},
			},
		},
	}
}

func (r *privAccessRuleResource) schemaUserAttributes() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:   "User attributes",
		Optional:      true,
		Computed:      true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
		Attributes: map[string]schema.Attribute{
			"risk_score": schema.SingleNestedAttribute{
				Description:   "User's risk score settings",
				Optional:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"category": schema.StringAttribute{
						Description:   "Risk score category",
						Required:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						Validators:    []validator.String{validators.RiscScoreCategory{}},
					},
					"operator": schema.StringAttribute{
						Description:   "Risk score operator",
						Required:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						Validators:    []validator.String{validators.RiskScoreOperator{}},
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
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"active_on": schema.StringAttribute{
				Description: "Type of a time range when a rule is active",
				Required:    true,
				Validators:  []validator.String{validators.PolicyActiveOnValidator{}},
			},
			"custom_recurring": schema.SingleNestedAttribute{
				Description: "Custom recurring time range that a rule is active",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"days": schema.ListAttribute{
						Description: "Days of the week",
						Required:    true,
						ElementType: types.StringType,
						Validators:  []validator.List{validators.DaysValidator{}},
					},
					"from": schema.StringAttribute{
						Description: "From time (12:34)",
						Required:    true,
					},
					"to": schema.StringAttribute{
						Description: "To time (12:34)",
						Required:    true,
					},
				},
			},
			"custom_timeframe": schema.SingleNestedAttribute{
				Description: "Custom one-time time range that a rule is active",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"from": schema.StringAttribute{
						Description: "From datetime (2006-01-02T15:04:05Z)",
						Required:    true,
					},
					"to": schema.StringAttribute{
						Description: "To datetime (2006-01-02T15:04:05Z)",
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
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"effective_from": schema.StringAttribute{
				Description: "Effective from",
				Optional:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "Expires at",
				Optional:    true,
			},
			"use_effective_from": schema.BoolAttribute{
				Description: "Use effective from",
				Optional:    true,
			},
			"use_expires_at": schema.BoolAttribute{
				Description: "Use expires at",
				Optional:    true,
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
				Description:   "Users",
				Optional:      true,
				Computed:      true,
				NestedObject:  schema.NestedAttributeObject{Attributes: parse.SchemaNameID("User")},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"user_groups": schema.ListNestedAttribute{
				Description:   "User groups",
				Optional:      true,
				Computed:      true,
				NestedObject:  schema.NestedAttributeObject{Attributes: parse.SchemaNameID("Group")},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
		},
		Validators: []validator.Object{validators.PrivAccPolicySourceValidator{}},
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

// Create private access policy rule
func (r *privAccessRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PrivateAccessRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleName := plan.Name.ValueString()
	input := cato_models.PrivateAccessAddRuleInput{
		Rule: &cato_models.PrivateAccessAddRuleDataInput{
			Action:           r.prepareAction(plan.Action),
			ActivePeriod:     r.prepareActivePeriod(ctx, plan.ActivePeriod, &resp.Diagnostics),
			Applications:     r.prepareApplications(ctx, plan.Applications, &resp.Diagnostics),
			ConnectionOrigin: r.prepareConnectionOrigins(ctx, plan.ConnectionOrigins, &resp.Diagnostics),
			Country:          r.prepareCountries(ctx, plan.Countries, &resp.Diagnostics),
			Description:      plan.Description.ValueString(),
			Device:           r.prepareDevice(ctx, plan.Devices, &resp.Diagnostics),
			Enabled:          plan.Enabled.ValueBool(),
			Name:             ruleName,
			Platform:         r.preparePlatforms(ctx, plan.Platforms, &resp.Diagnostics),
			Schedule:         r.prepareSchedule(ctx, plan.Schedule, &resp.Diagnostics),
			Source:           r.prepareSource(ctx, plan.Source, &resp.Diagnostics),
			Tracking:         r.prepareTracking(ctx, plan.Tracking, &resp.Diagnostics),
			UserAttributes:   r.prepareUserAttributes(ctx, plan.UserAttributes, &resp.Diagnostics),
		},
		At: &cato_models.PolicyRulePositionInput{
			Position: ptr(cato_models.PolicyRulePositionEnumLastInPolicy),
		},
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Call Cato API to create a new rule
	tflog.Debug(ctx, "PolicyPrivateAccessAddRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.catov2.PolicyPrivateAccessAddRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessAddRule", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	errMsg := fmt.Sprintf("failed to add private access rule '%s'", ruleName)
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	respRule := result.GetPolicy().GetPrivateAccess().GetAddRule()
	if respRule.Status != cato_models.PolicyMutationStatusSuccess {
		resp.Diagnostics.AddError(errMsg, "returned status: "+string(respRule.Status))
		for _, e := range respRule.Errors {
			resp.Diagnostics.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", *e.GetErrorMessage(), *e.GetErrorCode()))
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
	hydratedState, diags, hydrateErr := r.hydratePrivAccessRuleState(ctx, ruleID)
	if hydrateErr != nil {
		resp.Diagnostics.AddError("Error hydrating privateAccessRule state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read private access policy rule
func (r *privAccessRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PrivateAccessRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydratePrivAccessRuleState(ctx, state.ID.ValueString())
	if hydrateErr != nil {
		if errors.Is(hydrateErr, ErrPrivateAcccessRuleNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Private access rule %s not found, resource removed", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error hydrating privateAccessRule state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update private access policy rule
func (r *privAccessRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PrivateAccessRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PrivateAccessRuleModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("PolicyPrivateAccessUpdateRule: ID in unknown", "Rule ID is not set in TF state")
		return
	}

	input := cato_models.PrivateAccessUpdateRuleInput{
		ID: id,
		Rule: &cato_models.PrivateAccessUpdateRuleDataInput{
			Action:           r.prepareActionUpdate(plan.Action),
			ActivePeriod:     r.prepareActivePeriodUpdate(ctx, plan.ActivePeriod, &resp.Diagnostics),
			Applications:     r.prepareApplicationsUpdate(ctx, plan.Applications, &resp.Diagnostics),
			ConnectionOrigin: r.prepareConnectionOrigins(ctx, plan.ConnectionOrigins, &resp.Diagnostics),
			Country:          r.prepareCountries(ctx, plan.Countries, &resp.Diagnostics),
			Description:      parse.KnownStringPointer(plan.Description),
			Device:           r.prepareDevice(ctx, plan.Devices, &resp.Diagnostics),
			Enabled:          parse.KnownBoolPointer(plan.Enabled),
			Name:             parse.KnownStringPointer(plan.Name),
			Platform:         r.preparePlatforms(ctx, plan.Platforms, &resp.Diagnostics),
			Schedule:         r.prepareScheduleUpdate(ctx, plan.Schedule, &resp.Diagnostics),
			Source:           r.prepareSourceUpdate(ctx, plan.Source, &resp.Diagnostics),
			Tracking:         r.prepareTrackingUpdate(ctx, plan.Tracking, &resp.Diagnostics),
			UserAttributes:   r.prepareUserAttributesUpdate(ctx, plan.UserAttributes, &resp.Diagnostics),
		},
	}

	tflog.Debug(ctx, "PolicyPrivateAccessUpdateRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.catov2.PolicyPrivateAccessUpdateRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessUpdateRule", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	errMsg := fmt.Sprintf("failed to update private access rule '%s'", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	res := result.GetPolicy().GetPrivateAccess().GetUpdateRule()
	if *res.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		errors := res.GetErrors()
		if len(errors) == 0 {
			resp.Diagnostics.AddError(errMsg, "returned status: "+string(*res.GetStatus()))
			return
		}
		for _, e := range errors {
			resp.Diagnostics.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", *e.GetErrorMessage(), *e.GetErrorCode()))
		}
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydratePrivAccessRuleState(ctx, id)
	if hydrateErr != nil {
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("Error hydrating private access policy rule state", hydrateErr.Error())
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete private access policy rule
func (r *privAccessRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PrivateAccessRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := cato_models.PrivateAccessRemoveRuleInput{ID: state.ID.ValueString()}

	// Call Cato API to delete a connector
	tflog.Debug(ctx, "PolicyPrivateAccessDeleteRule", map[string]interface{}{"request": utils.InterfaceToJSONString(input)})
	result, err := r.client.catov2.PolicyPrivateAccessDeleteRule(ctx, r.client.AccountId, input)
	tflog.Debug(ctx, "PolicyPrivateAccessDeleteRule", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	errMsg := fmt.Sprintf("failed to delete private access rule '%s'", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errMsg, err.Error())
		return
	}
	res := result.GetPolicy().GetPrivateAccess().GetRemoveRule()
	if *res.GetStatus() != cato_models.PolicyMutationStatusSuccess {
		errors := res.GetErrors()
		if len(errors) == 0 {
			resp.Diagnostics.AddError(errMsg, "returned status: "+string(*res.GetStatus()))
			return
		}
		for _, e := range errors {
			resp.Diagnostics.AddError(errMsg, fmt.Sprintf("ERROR: %v [%v]", *e.GetErrorMessage(), *e.GetErrorCode()))
		}
		return
	}
}

func (r *privAccessRuleResource) parseSource(ctx context.Context, src cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Source,
	diags *diag.Diagnostics,
) types.Object {
	tfSource := Source{
		Users:      parse.ParseIDRefList(ctx, src.User, diags),
		UserGroups: parse.ParseIDRefList(ctx, src.UsersGroup, diags),
	}
	obj, diag := types.ObjectValueFrom(ctx, SourceTypes, tfSource)
	diags.Append(diag...)
	return obj
}

func (r *privAccessRuleResource) prepareSource(ctx context.Context, src types.Object, diags *diag.Diagnostics) *cato_models.PrivateAccessPolicySourceInput {
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
	userInput = parse.PrepareIDRefList[cato_models.UserRefInput](ctx, tfSource.Users, diags, "source.users")
	if diags.HasError() {
		return nil
	}

	// group list
	groupInput = parse.PrepareIDRefList[cato_models.UsersGroupRefInput](ctx, tfSource.UserGroups, diags, "source.groups")
	if diags.HasError() {
		return nil
	}

	return &cato_models.PrivateAccessPolicySourceInput{
		User:       userInput,
		UsersGroup: groupInput,
	}
}

func (r *privAccessRuleResource) prepareSourceUpdate(ctx context.Context, src types.Object, diags *diag.Diagnostics) *cato_models.PrivateAccessPolicySourceUpdateInput {
	upd := r.prepareSource(ctx, src, diags)
	if upd == nil {
		return nil
	}
	return &cato_models.PrivateAccessPolicySourceUpdateInput{User: upd.User, UsersGroup: upd.UsersGroup}
}

func (r *privAccessRuleResource) prepareApplications(ctx context.Context, apps types.List, diags *diag.Diagnostics) *cato_models.PrivateAccessPolicyApplicationInput {
	if !utils.HasValue(apps) {
		return nil
	}
	applicationInput := parse.PrepareIDRefList[cato_models.PrivateApplicationRefInput](ctx, apps, diags, "applications")
	if diags.HasError() {
		return nil
	}

	return &cato_models.PrivateAccessPolicyApplicationInput{Application: applicationInput}
}

func (r *privAccessRuleResource) prepareApplicationsUpdate(ctx context.Context, apps types.List, diags *diag.Diagnostics) *cato_models.PrivateAccessPolicyApplicationUpdateInput {
	upd := r.prepareApplications(ctx, apps, diags)
	if upd == nil {
		return nil
	}
	return &cato_models.PrivateAccessPolicyApplicationUpdateInput{Application: upd.Application}
}

func (r *privAccessRuleResource) prepareTracking(ctx context.Context, t types.Object, diags *diag.Diagnostics) *cato_models.PolicyTrackingInput {
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
	sdkTracking.Alert.MailingList = parse.PrepareIDRefList[cato_models.SubscriptionMailingListRefInput](ctx, tfAlert.MailingList, diags, "tracking.alert.mailing_list")
	if diags.HasError() {
		return nil
	}

	// Subscription groups
	sdkTracking.Alert.SubscriptionGroup = parse.PrepareIDRefList[cato_models.SubscriptionGroupRefInput](ctx, tfAlert.SubscriptionGroup, diags, "tracking.alert.subscription_group")
	if diags.HasError() {
		return nil
	}

	// Webhooks
	sdkTracking.Alert.Webhook = parse.PrepareIDRefList[cato_models.SubscriptionWebhookRefInput](ctx, tfAlert.Webhook, diags, "tracking.alert.webhook")
	if diags.HasError() {
		return nil
	}

	return &sdkTracking
}
func (r *privAccessRuleResource) prepareTrackingUpdate(ctx context.Context, t types.Object, diags *diag.Diagnostics) *cato_models.PolicyTrackingUpdateInput {
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
			Enabled: ptr(upd.Alert.Enabled),
		},
	}
}

func (r *privAccessRuleResource) prepareUserAttributes(ctx context.Context, uas types.Object, diags *diag.Diagnostics) *cato_models.PrivateAccessUserAttributesInput {
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

func (r *privAccessRuleResource) prepareUserAttributesUpdate(ctx context.Context, uas types.Object, diags *diag.Diagnostics) *cato_models.PrivateAccessUserAttributesUpdateInput {
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

func (r *privAccessRuleResource) prepareSchedule(ctx context.Context, sch types.Object, diags *diag.Diagnostics) *cato_models.PolicyScheduleInput {
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

func (r *privAccessRuleResource) prepareScheduleUpdate(ctx context.Context, sch types.Object, diags *diag.Diagnostics) *cato_models.PolicyScheduleUpdateInput {
	upd := r.prepareSchedule(ctx, sch, diags)
	if upd == nil {
		return nil
	}
	return &cato_models.PolicyScheduleUpdateInput{
		ActiveOn: &upd.ActiveOn,
	}
}

func (r *privAccessRuleResource) prepareActivePeriod(ctx context.Context, ap types.Object, diags *diag.Diagnostics) *cato_models.PolicyRuleActivePeriodInput {
	if ap.IsUnknown() || ap.IsNull() {
		return &cato_models.PolicyRuleActivePeriodInput{}
	}

	var tfPeriod PolicyRuleActivePeriod
	diags.Append(ap.As(ctx, &tfPeriod, basetypes.ObjectAsOptions{})...)

	sdkPeriod := cato_models.PolicyRuleActivePeriodInput{
		EffectiveFrom:    parse.KnownStringPointer(tfPeriod.EffectiveFrom),
		ExpiresAt:        parse.KnownStringPointer(tfPeriod.ExpiresAt),
		UseEffectiveFrom: tfPeriod.UseEffectiveFrom.ValueBool(),
		UseExpiresAt:     tfPeriod.UseExpiresAt.ValueBool(),
	}

	return &sdkPeriod
}

func (r *privAccessRuleResource) prepareActivePeriodUpdate(ctx context.Context, ap types.Object, diags *diag.Diagnostics) *cato_models.PolicyRuleActivePeriodUpdateInput {
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

func (r *privAccessRuleResource) preparePlatforms(ctx context.Context, platforms types.List, diags *diag.Diagnostics) []cato_models.OperatingSystem {
	return parse.PrepareStrings[cato_models.OperatingSystem](ctx, platforms, diags, "rule.platforms")
}

func (r *privAccessRuleResource) prepareCountries(ctx context.Context, countries types.List, diags *diag.Diagnostics) []*cato_models.CountryRefInput {
	return parse.PrepareIDRefList[cato_models.CountryRefInput](ctx, countries, diags, "rule.countries")
}

func (r *privAccessRuleResource) prepareConnectionOrigins(ctx context.Context, os types.List, diags *diag.Diagnostics) []cato_models.PrivateAccessPolicyOriginEnum {
	return parse.PrepareStrings[cato_models.PrivateAccessPolicyOriginEnum](ctx, os, diags, "rule.connection_origins")
}

func (r *privAccessRuleResource) prepareAction(action types.String) *cato_models.PrivateAccessPolicyActionInput {
	return &cato_models.PrivateAccessPolicyActionInput{Action: cato_models.PrivateAccessPolicyActionEnum(action.ValueString())}
}
func (r *privAccessRuleResource) prepareActionUpdate(action types.String) *cato_models.PrivateAccessPolicyActionUpdateInput {
	return &cato_models.PrivateAccessPolicyActionUpdateInput{Action: ptr(cato_models.PrivateAccessPolicyActionEnum(action.ValueString()))}
}

func (r *privAccessRuleResource) prepareDevice(ctx context.Context, devs types.List, diags *diag.Diagnostics) []*cato_models.DeviceProfileRefInput {
	return parse.PrepareIDRefList[cato_models.DeviceProfileRefInput](ctx, devs, diags, "rule.devices")
}

func (r *privAccessRuleResource) parseTracking(ctx context.Context, tr cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Tracking,
	diags *diag.Diagnostics,
) types.Object {
	var diag diag.Diagnostics

	// Prepare Tracking.Event object
	tfEvent := PolicyRuleTrackingEvent{
		Enabled: types.BoolValue(tr.Event.Enabled),
	}
	eventObj, diag := types.ObjectValueFrom(ctx, PolicyRuleTrackingEventTypes, tfEvent)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleTrackingTypes)
	}

	// Prepare Tracking.Alert object
	mailingList := parse.ParseIDRefList(ctx, tr.Alert.MailingList, diags)
	subscriptionGroup := parse.ParseIDRefList(ctx, tr.Alert.SubscriptionGroup, diags)
	webHook := parse.ParseIDRefList(ctx, tr.Alert.Webhook, diags)
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
	alertObj, diag := types.ObjectValueFrom(ctx, PolicyRuleTrackingAlertTypes, tfAlert)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleTrackingTypes)
	}

	// Prepare Tracking object
	tfTracking := PolicyRuleTracking{
		Event: eventObj,
		Alert: alertObj,
	}
	trackingObj, diag := types.ObjectValueFrom(ctx, PolicyRuleTrackingTypes, tfTracking)
	diags.Append(diag...)

	return trackingObj
}

func (r *privAccessRuleResource) parseUserAttributes(ctx context.Context, ua cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_UserAttributes,
	diags *diag.Diagnostics,
) types.Object {
	var diag diag.Diagnostics

	// Prepare UserAttributes.RiskScore object
	tfRiskScore := RiskScore{
		Category: types.StringValue(string(ua.RiskScore.Category)),
		Operator: types.StringValue(string(ua.RiskScore.Operator)),
	}
	riskScoreObj, diag := types.ObjectValueFrom(ctx, RiskScoreTypes, tfRiskScore)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(UserAttributesTypes)
	}

	// Prepare UserAttributes object
	tfUserAttributes := UserAttributes{
		RiskScore: riskScoreObj,
	}
	userAttrObj, diag := types.ObjectValueFrom(ctx, UserAttributesTypes, tfUserAttributes)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleTrackingTypes)
	}

	return userAttrObj
}

func (r *privAccessRuleResource) parsePolicySchedule(ctx context.Context, sch cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Schedule,
	diags *diag.Diagnostics,
) types.Object {
	var diag diag.Diagnostics

	// Prepare PolicySchedule.PolicyCustomRecurring object
	var recurringObj types.Object = types.ObjectNull(PolicyCustomRecurringTypes)
	if sch.CustomRecurring != nil {
		tfRecurring := PolicyCustomRecurring{
			Days: parse.ParseStringList(ctx, sch.CustomRecurring.Days, diags),
			From: types.StringValue(string(sch.CustomRecurring.From)),
			To:   types.StringValue(string(sch.CustomRecurring.To)),
		}
		recurringObj, diag = types.ObjectValueFrom(ctx, PolicyCustomRecurringTypes, tfRecurring)
		diags.Append(diag...)
		if diags.HasError() {
			return types.ObjectNull(PolicyScheduleTypes)
		}
	}

	// Prepare PolicySchedule.PolicyCustomTimeframe object
	var timeframeObj types.Object = types.ObjectNull(PolicyCustomTimeframeTypes)
	if sch.CustomTimeframe != nil {
		tfTimeframe := PolicyCustomTimeframe{
			From: types.StringValue(string(sch.CustomTimeframe.From)),
			To:   types.StringValue(string(sch.CustomTimeframe.To)),
		}
		timeframeObj, diag = types.ObjectValueFrom(ctx, PolicyCustomTimeframeTypes, tfTimeframe)
		diags.Append(diag...)
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
	scheduleObj, diag := types.ObjectValueFrom(ctx, PolicyScheduleTypes, tfSchedule)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(PolicyScheduleTypes)
	}

	return scheduleObj
}

func (r *privAccessRuleResource) parsePolicyActivePeriod(ctx context.Context, ap cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_ActivePeriod,
	diags *diag.Diagnostics,
) types.Object {
	var diag diag.Diagnostics

	// Prepare Active Period
	tfActivePeriod := PolicyRuleActivePeriod{
		EffectiveFrom:    types.StringPointerValue(ap.EffectiveFrom),
		ExpiresAt:        types.StringPointerValue(ap.ExpiresAt),
		UseEffectiveFrom: types.BoolValue(ap.UseEffectiveFrom),
		UseExpiresAt:     types.BoolValue(ap.UseExpiresAt),
	}
	periodObj, diag := types.ObjectValueFrom(ctx, PolicyRuleActivePeriodTypes, tfActivePeriod)
	diags.Append(diag...)
	if diags.HasError() {
		return types.ObjectNull(PolicyRuleActivePeriodTypes)
	}

	return periodObj
}

// hydratePrivAccessRuleState fetches the current state of a privAccessRule from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessRuleResource) hydratePrivAccessRuleState(ctx context.Context, ruleID string) (*PrivateAccessRuleModel, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	// Call Cato API to get the policy
	result, err := r.client.catov2.PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "PolicyReadPrivateAccessPolicy", map[string]interface{}{"response": utils.InterfaceToJSONString(result)})
	if err != nil {
		return nil, nil, err
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
			Action:            types.StringValue(string(apiRule.Action.Action)),
			ActivePeriod:      r.parsePolicyActivePeriod(ctx, apiRule.ActivePeriod, &diags),
			Applications:      parse.ParseIDRefList(ctx, apiRule.Applications.Application, &diags),
			ConnectionOrigins: parse.ParseStringList(ctx, apiRule.ConnectionOrigin, &diags),
			Countries:         parse.ParseIDRefList(ctx, apiRule.Country, &diags),
			Description:       types.StringValue(apiRule.Description),
			Devices:           parse.ParseIDRefList(ctx, apiRule.Device, &diags),
			Enabled:           types.BoolValue(apiRule.Enabled),
			ID:                types.StringValue(apiRule.ID),
			Name:              types.StringValue(apiRule.Name),
			Platforms:         parse.ParseStringList(ctx, apiRule.Platform, &diags),
			Schedule:          r.parsePolicySchedule(ctx, apiRule.Schedule, &diags),
			Source:            r.parseSource(ctx, apiRule.Source, &diags),
			Tracking:          r.parseTracking(ctx, apiRule.Tracking, &diags),
			UserAttributes:    r.parseUserAttributes(ctx, apiRule.UserAttributes, &diags),
		}
		break
	}

	if state == nil {
		return nil, diags, ErrPrivateAcccessRuleNotFound
	}
	if diags.HasError() {
		return nil, diags, ErrAPIResponseParse
	}

	return state, nil, nil
}
