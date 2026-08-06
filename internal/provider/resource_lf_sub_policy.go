package provider

import (
	"context"
	"errors"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
)

var (
	_ resource.Resource                = &lfSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &lfSubPolicyResource{}
	_ resource.ResourceWithImportState = &lfSubPolicyResource{}

	ErrLanRuleNotFound = errors.New("lan rule not found")
)

func NewLfSubPolicyResource() resource.Resource {
	return &lfSubPolicyResource{}
}

type lfSubPolicyResource struct {
	client *catoClientData
}

type fromtoer interface {
	GetFrom() string
	GetTo() string
}

func (r *lfSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lf_sub_policy"
}

func (r *lfSubPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: "The `cato_lf_sub_policy` resource manages a LAN Firewall sub-policy " +
			"(a nested policy scoped by a SUB_POLICY_SCOPE rule).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Sub-policy ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "Sub-policy name. Changing this forces replacement.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Description:   "Sub-policy description. Changing this forces replacement.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"scope_rule_id": schema.StringAttribute{
				Description:   "ID of the underlying SUB_POLICY_SCOPE rule.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"at": schema.SingleNestedAttribute{
				Description:   "Position of the sub-policy scope within the LAN Firewall policy.",
				Required:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule.",
						Required:    true,
					},
					"ref": schema.StringAttribute{
						Description: "Identifier of the object relative to which the position is defined.",
						Optional:    true,
					},
				},
			},
			"scope": r.policyScopeSchema(),
		},
	}
}

func (r *lfSubPolicyResource) policyScopeSchema() schema.SingleNestedAttribute {
	scopeAttr := (&socketLanNetworkRuleResource{}).lanRuleSchema()
	scopeAttr.Description = "Policy scope attributes"
	scopeAttr.Required = true
	delete(scopeAttr.Attributes, "transport")
	scopeAttr.Attributes["description"] = schema.StringAttribute{
		Description:   "API-managed scope description, synchronized with the sub-policy description.",
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
	scopeAttr.Attributes["name"] = schema.StringAttribute{
		Description:   "API-managed scope name, synchronized with the sub-policy name.",
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}

	return scopeAttr
}

func (r *lfSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *lfSubPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *lfSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *lfSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LanFirewallSubPolicy
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, diags, hydrateErr := r.hydrateLfSubPolicy(ctx, state.ID.ValueString(), &state)
	if hydrateErr != nil {
		if errors.Is(hydrateErr, ErrLanRuleNotFound) {
			tflog.Warn(ctx, fmt.Sprintf("Lan policy rule %s not found, resource removed", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error hydrating lan sub-policy state", hydrateErr.Error())
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lfSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *lfSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// hydrateLfSubPolicy fetches the current state of a lan firewall sub-policy from the API
func (r *lfSubPolicyResource) hydrateLfSubPolicy(ctx context.Context, subPolicyID string,
	cfgOrState *LanFirewallSubPolicy,
) (*LanFirewallSubPolicy, diag.Diagnostics, error) {
	var diags diag.Diagnostics

	// Call Cato API to get the policy
	result, err := r.client.catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		return nil, nil, err
	}

	var state *LanFirewallSubPolicy

	// Map API response to LanFirewallSubPolicy
	policy := result.GetPolicy().GetSocketLan().GetPolicy()
	for _, polRule := range policy.Rules {
		if polRule.GetSubPolicy().ID != subPolicyID {
			continue
		}
		apiRule := polRule.Rule
		state = &LanFirewallSubPolicy{
			ID:          types.StringValue(subPolicyID),
			Name:        types.StringValue(apiRule.Name),
			Description: types.StringValue(apiRule.Description),
			At:          cfgOrState.At,
			ScopeRuleID: types.StringValue(apiRule.ID),
			Scope:       r.parseRuleScope(ctx, apiRule, &diags),
		}
		break
	}

	if state == nil {
		return nil, diags, ErrLanRuleNotFound
	}
	if diags.HasError() {
		return nil, diags, ErrAPIResponseParse
	}

	return state, nil, nil
}

// parseRuleScope pares the rule scope from the API response and returns terraform LanFirewallSubPolicyScope object
func (r *lfSubPolicyResource) parseRuleScope(ctx context.Context,
	rule cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule, diags *diag.Diagnostics,
) types.Object {
	// Prepare LanFirewallSubPolicyScope object
	tfScope := LanFirewallSubPolicyScope{
		Description: types.StringValue(rule.Description),
		Destination: r.parseDestination(ctx, rule.Destination, diags),
		Direction:   types.StringValue(string(rule.Direction)),
		Enabled:     types.BoolValue(rule.Enabled),
		ID:          types.StringValue(rule.ID),
		Name:        types.StringValue(rule.Name),
		NAT:         r.parseNat(ctx, rule.Nat, diags),
		Service:     r.parseService(ctx, rule.Service, diags),
		Site:        r.parseSite(ctx, rule.Site, diags),
		Source:      r.parseSource(ctx, rule.Source, diags),
	}
	subPolicyObj, objDiags := types.ObjectValueFrom(ctx, LanFirewallSubPolicyScopeTypes, tfScope)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(LanFirewallSubPolicyScopeTypes)
	}

	return subPolicyObj
}

// parseSource parses the source of a rule into a SocketLanDource object
func (r *lfSubPolicyResource) parseSource(ctx context.Context,
	dst cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Source, diags *diag.Diagnostics,
) types.Object {
	tfSocketLanSource := SocketLanSource{
		Vlan:              parse.Int64List(ctx, dst.GetVlan(), diags),
		IP:                parse.StringList(ctx, dst.GetIP(), diags),
		Subnet:            parse.StringList(ctx, dst.GetSubnet(), diags),
		IPRange:           FromToList(ctx, dst.GetIPRange(), diags),
		Host:              parse.IDRefSet(ctx, dst.GetHost(), diags),
		Group:             parse.IDRefSet(ctx, dst.GetGroup(), diags),
		SystemGroup:       parse.IDRefSet(ctx, dst.GetSystemGroup(), diags),
		NetworkInterface:  parse.IDRefSet(ctx, dst.GetNetworkInterface(), diags),
		GlobalIPRange:     parse.IDRefSet(ctx, dst.GetGlobalIPRange(), diags),
		FloatingSubnet:    parse.IDRefSet(ctx, dst.GetFloatingSubnet(), diags),
		SiteNetworkSubnet: parse.IDRefSet(ctx, dst.GetSiteNetworkSubnet(), diags),
	}
	ruleSourceObj, objDiags := types.ObjectValueFrom(ctx, SocketLanSourceAttrTypes, tfSocketLanSource)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(SocketLanSourceAttrTypes)
	}

	return ruleSourceObj
}

// parseDestination parses the destination of a rule into a SocketLanDestination object
func (r *lfSubPolicyResource) parseDestination(ctx context.Context,
	dst cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Destination, diags *diag.Diagnostics,
) types.Object {
	tfSocketLanDestination := SocketLanDestination{
		Vlan:              parse.Int64List(ctx, dst.GetVlan(), diags),
		IP:                parse.StringList(ctx, dst.GetIP(), diags),
		Subnet:            parse.StringList(ctx, dst.GetSubnet(), diags),
		IPRange:           FromToList(ctx, dst.GetIPRange(), diags),
		Host:              parse.IDRefSet(ctx, dst.GetHost(), diags),
		Group:             parse.IDRefSet(ctx, dst.GetGroup(), diags),
		SystemGroup:       parse.IDRefSet(ctx, dst.GetSystemGroup(), diags),
		NetworkInterface:  parse.IDRefSet(ctx, dst.GetNetworkInterface(), diags),
		GlobalIPRange:     parse.IDRefSet(ctx, dst.GetGlobalIPRange(), diags),
		FloatingSubnet:    parse.IDRefSet(ctx, dst.GetFloatingSubnet(), diags),
		SiteNetworkSubnet: parse.IDRefSet(ctx, dst.GetSiteNetworkSubnet(), diags),
	}
	ruleDestinationObj, objDiags := types.ObjectValueFrom(ctx, SocketLanDestinationAttrTypes, tfSocketLanDestination)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(SocketLanDestinationAttrTypes)
	}

	return ruleDestinationObj
}

// FromToList parses a slice of ip ranges (from-to) into a types.List of FromToAttrTypes
func FromToList[T fromtoer](ctx context.Context, fts []T, diags *diag.Diagnostics) types.List {
	// null value
	if fts == nil {
		return types.ListNull(FromToObjectType)
	}

	// existing empty list
	if len(fts) == 0 {
		val, valueDiag := types.ListValue(FromToObjectType, nil)
		diags.Append(valueDiag...)
		return val
	}

	// make []FromToObject
	fromToSlice := make([]types.Object, 0, len(fts))
	for _, ft := range fts {
		tfFromTo := FromTo{
			From: types.StringValue(ft.GetFrom()),
			To:   types.StringValue(ft.GetTo()),
		}
		ftObj, objDiags := types.ObjectValueFrom(ctx, FromToAttrTypes, tfFromTo)
		diags.Append(objDiags...)
		fromToSlice = append(fromToSlice, ftObj)
	}
	// convert to types.List
	fromToList, valueDiag := types.ListValueFrom(ctx, FromToObjectType, fromToSlice)
	diags.Append(valueDiag...)

	return fromToList
}

func (r *lfSubPolicyResource) parseNat(ctx context.Context,
	nat cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Nat, diags *diag.Diagnostics,
) types.Object {
	tfNat := PolicyNatSettings{
		Enabled: types.BoolValue(nat.Enabled),
		NatType: types.StringValue(string(nat.NatType)),
	}
	natObj, objDiags := types.ObjectValueFrom(ctx, PolicyNatSettingsTypes, tfNat)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyNatSettingsTypes)
	}

	return natObj
}

func (r *lfSubPolicyResource) parseService(ctx context.Context,
	svc cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service, diags *diag.Diagnostics,
) types.Object {
	tfSvc := PolicyService{
		Custom: r.parseCustomService(ctx, svc.Custom, diags),
		Simple: r.parseSimpleService(ctx, svc.Simple, diags),
	}
	svcObj, objDiags := types.ObjectValueFrom(ctx, PolicyServiceTypes, tfSvc)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyServiceTypes)
	}

	return svcObj
}

func (r *lfSubPolicyResource) parseSimpleService(ctx context.Context,
	svcs []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service_Simple, diags *diag.Diagnostics,
) types.Set {
	serviceSlice := make([]types.Object, 0, len(svcs))
	for _, svc := range svcs {
		tfSimpleService := SimpleService{
			Name: types.StringValue(svc.GetName().String()),
		}
		simpleServiceObj, objDiags := types.ObjectValueFrom(ctx, SimpleServiceAttrTypes, tfSimpleService)
		diags.Append(objDiags...)
		serviceSlice = append(serviceSlice, simpleServiceObj)
	}
	serviceList, valueDiag := types.SetValueFrom(ctx, SimpleServiceObjectType, serviceSlice)
	diags.Append(valueDiag...)

	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: SimpleServiceAttrTypes})
	}

	return serviceList
}

func (r *lfSubPolicyResource) parseCustomService(ctx context.Context,
	svcs []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service_Custom, diags *diag.Diagnostics,
) types.List {
	serviceSlice := make([]types.Object, 0, len(svcs))
	for _, svc := range svcs {
		tfPortRange := FromTo{
			From: types.StringPointerValue((*string)(svc.PortRange.GetFrom())),
			To:   types.StringPointerValue((*string)(svc.PortRange.GetTo())),
		}
		portRangeObj, objDiags := types.ObjectValueFrom(ctx, FromToAttrTypes, tfPortRange)
		diags.Append(objDiags...)

		tfCustomService := PolicyCustomService{
			PortRange: portRangeObj,
			Port:      parse.StringList(ctx, svc.Port, diags),
			Protocol:  types.StringValue(string(svc.Protocol)),
		}
		customServiceObj, objDiags := types.ObjectValueFrom(ctx, PolicyCustomServiceTypes, tfCustomService)
		diags.Append(objDiags...)
		serviceSlice = append(serviceSlice, customServiceObj)
	}
	serviceList, valueDiag := types.ListValueFrom(ctx, CustomServiceObjectType, serviceSlice)
	diags.Append(valueDiag...)

	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: PolicyCustomServiceTypes})
	}

	return serviceList
}

func (r *lfSubPolicyResource) parseSite(ctx context.Context,
	site cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Site, diags *diag.Diagnostics,
) types.Object {
	tfSite := PolicySite{
		Group: parse.IDRefSet(ctx, site.GetGroup(), diags),
		Site:  parse.IDRefSet(ctx, site.GetSite(), diags),
	}
	siteObj, objDiags := types.ObjectValueFrom(ctx, PolicySiteTypes, tfSite)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicySiteTypes)
	}

	return siteObj
}
