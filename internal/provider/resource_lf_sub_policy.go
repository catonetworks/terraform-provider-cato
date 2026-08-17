package provider

import (
	"context"
	"errors"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &lfSubPolicyResource{}
	_ resource.ResourceWithConfigure   = &lfSubPolicyResource{}
	_ resource.ResourceWithImportState = &lfSubPolicyResource{}

	ErrLanRuleNotFound = errors.New("lan rule not found")
)

// NewLfSubPolicyResource creates a LAN Firewall sub-policy resource instance.
func NewLfSubPolicyResource() resource.Resource {
	return &lfSubPolicyResource{}
}

type lfSubPolicyResource struct {
	client *catoClientData
}

type requiresReplaceForConfiguredPosition struct{}

func (requiresReplaceForConfiguredPosition) Description(_ context.Context) string {
	return "Replace when an already-configured sub-policy position changes."
}

func (m requiresReplaceForConfiguredPosition) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (requiresReplaceForConfiguredPosition) PlanModifyObject(
	_ context.Context,
	req planmodifier.ObjectRequest,
	resp *planmodifier.ObjectResponse,
) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if req.PlanValue.IsUnknown() {
		resp.RequiresReplace = true
		return
	}
	if !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}

type fromtoer interface {
	GetFrom() string
	GetTo() string
}

// Metadata sets the Terraform resource type name for LAN Firewall sub-policies.
func (r *lfSubPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lf_sub_policy"
}

// Schema defines the Terraform schema for a LAN Firewall sub-policy.
func (r *lfSubPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Optional:      true,
				PlanModifiers: []planmodifier.Object{requiresReplaceForConfiguredPosition{}},
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

// policyScopeSchema derives the sub-policy scope schema from the LAN rule schema.
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

// Configure stores the configured Cato API client for resource operations.
func (r *lfSubPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

// ImportState imports a LAN Firewall sub-policy by its ID.
func (r *lfSubPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create adds a LAN Firewall sub-policy and hydrates its Terraform state from the API.
func (r *lfSubPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LanFirewallSubPolicy
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !utils.HasValue(plan.At) {
		resp.Diagnostics.AddAttributeError(
			path.Root("at"),
			"Missing LAN sub-policy position",
			"The at attribute must be configured when creating a LAN sub-policy.",
		)
		return
	}

	input := cato_models.SocketLanAddSubPolicyInput{
		At:     r.prepareAt(ctx, plan.At, &resp.Diagnostics),
		Policy: r.preparePolicy(plan.Name, plan.Description),
		Scope:  r.prepareScope(ctx, plan.Scope, plan.Name, plan.Description, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.catov2.PolicySocketLanAddSubPolicy(ctx, input, r.client.AccountId)
	subpol := result.GetPolicy().GetSocketLan().GetAddSubPolicy()
	if utils.CheckAPIErrors(err, subpol.GetErrors(),
		fmt.Sprintf("failed to add LAN sub-policy '%s'", plan.Name), &resp.Diagnostics) {
		return
	}

	policyID := subpol.GetPolicy().GetID()
	if policyID == "" {
		resp.Diagnostics.AddError("Cato API PolicySocketLanAddSubPolicy error", "policy ID is empty")
		return
	}

	// Set the ID from the response
	plan.ID = types.StringValue(policyID)

	// publish the changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, plan)

	// Hydrate state from API
	hydratedState, notFound := r.hydrateLfSubPolicy(ctx, policyID, &plan, &resp.Diagnostics)
	if notFound {
		resp.Diagnostics.AddError("failed to create sub policy", "sub-policy not found in API response")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
}

// Update modifies the scope rule of a LAN Firewall sub-policy and refreshes its state.
func (r *lfSubPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LanFirewallSubPolicy
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tfScope LanFirewallSubPolicyScope
	if utils.CheckErr(&resp.Diagnostics, plan.Scope.As(ctx, &tfScope, basetypes.ObjectAsOptions{})) {
		return
	}

	input := cato_models.SocketLanUpdateRuleInput{
		ID: plan.ScopeRuleID.ValueString(),
		Rule: &cato_models.SocketLanUpdateRuleDataInput{
			Description: plan.Description.ValueStringPointer(),
			Destination: r.prepareDestinationUpdate(ctx, tfScope.Destination, &resp.Diagnostics),
			Direction:   (*cato_models.SocketLanDirection)(tfScope.Direction.ValueStringPointer()),
			Enabled:     tfScope.Enabled.ValueBoolPointer(),
			Name:        plan.Name.ValueStringPointer(),
			Nat:         r.prepareNatUpdate(ctx, tfScope.NAT, &resp.Diagnostics),
			Service:     r.prepareServiceUpdate(ctx, tfScope.Service, &resp.Diagnostics),
			Site:        r.prepareSiteUpdate(ctx, tfScope.Site, &resp.Diagnostics),
			Source:      r.prepareSourceUpdate(ctx, tfScope.Source, &resp.Diagnostics),
			Transport:   ptr(cato_models.SocketLanTransportTypeLan),
		},
	}

	result, err := r.client.catov2.PolicySocketLanUpdateRule(ctx, nil, input, r.client.AccountId)
	subpol := result.GetPolicy().GetSocketLan().GetUpdateRule()
	if utils.CheckAPIErrors(err, subpol.GetErrors(),
		fmt.Sprintf("failed to update LAN sub-policy '%s'", plan.Name), &resp.Diagnostics) {
		return
	}

	// publish the changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, notFound := r.hydrateLfSubPolicy(ctx, plan.ID.ValueString(), &plan, &resp.Diagnostics)
	if notFound {
		resp.Diagnostics.AddError("failed to update sub policy", "sub-policy not found in API response")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
}

// Delete removes a LAN Firewall sub-policy through the Cato API.
func (r *lfSubPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LanFirewallSubPolicy
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := cato_models.SocketLanRemoveSubPolicyInput{
		Ref: &cato_models.SocketLanPolicyRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.ID.ValueString(),
		},
	}

	// Call Cato API to delete the subpolicy
	res, err := r.client.catov2.PolicySocketLanRemoveSubPolicy(ctx, nil, input, r.client.AccountId)
	if utils.CheckAPIErrors(err, res.GetPolicy().GetSocketLan().GetRemoveSubPolicy().GetErrors(),
		fmt.Sprintf("failed to delete lan sub-policy '%s'", state.Name.ValueString()), &resp.Diagnostics) {
		return
	}

	// publish the changes
	r.publish(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes a LAN Firewall sub-policy from the Cato API and removes missing resources from state.
func (r *lfSubPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LanFirewallSubPolicy
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate state from API
	hydratedState, notFound := r.hydrateLfSubPolicy(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)
	if notFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// prepareAt converts the Terraform rule position into an API position input.
func (r *lfSubPolicyResource) prepareAt(ctx context.Context, at types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyRulePositionInput {
	if !utils.HasValue(at) {
		return nil
	}
	var tfPosition PolicyRulePositionInput
	if utils.CheckErr(diags, at.As(ctx, &tfPosition, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.PolicyRulePositionInput{
		Position: (*cato_models.PolicyRulePositionEnum)(tfPosition.Position.ValueStringPointer()),
		Ref:      tfPosition.Ref.ValueStringPointer(),
	}
}

// preparePolicy builds the API input containing the sub-policy name and description.
func (r *lfSubPolicyResource) preparePolicy(name, desc types.String) *cato_models.SocketLanAddSubPolicyDataInput {
	return &cato_models.SocketLanAddSubPolicyDataInput{
		Description: desc.ValueString(),
		Name:        name.ValueString(),
	}
}

// prepareScope converts the Terraform scope rule into an API add-rule input.
func (r *lfSubPolicyResource) prepareScope(ctx context.Context, scope types.Object, name, desc types.String, diags *diag.Diagnostics,
) *cato_models.SocketLanAddRuleDataInput {
	if !utils.HasValue(scope) {
		return nil
	}
	var tfScope LanFirewallSubPolicyScope
	if utils.CheckErr(diags, scope.As(ctx, &tfScope, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.SocketLanAddRuleDataInput{
		Description: desc.ValueString(),
		Destination: r.prepareDestination(ctx, tfScope.Destination, diags),
		Direction:   cato_models.SocketLanDirection(tfScope.Direction.ValueString()),
		Enabled:     tfScope.Enabled.ValueBool(),
		Name:        name.ValueString(),
		Nat:         r.prepareNat(ctx, tfScope.NAT, diags),
		Service:     r.prepareService(ctx, tfScope.Service, diags),
		Site:        r.prepareSite(ctx, tfScope.Site, diags),
		Source:      r.prepareSource(ctx, tfScope.Source, diags),
		Transport:   cato_models.SocketLanTransportTypeLan,
	}
}

// prepareSite converts Terraform site and group references into an API site input.
func (r *lfSubPolicyResource) prepareSite(ctx context.Context, site types.Object, diags *diag.Diagnostics) *cato_models.SocketLanSiteInput {
	if !utils.HasValue(site) {
		return nil
	}
	var tfSite SocketLanSite
	if utils.CheckErr(diags, site.As(ctx, &tfSite, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanSiteInput{
		Group: parse.PrepareIDRefSet[cato_models.GroupRefInput](ctx, tfSite.Group, diags),
		Site:  parse.PrepareIDRefSet[cato_models.SiteRefInput](ctx, tfSite.Site, diags),
	}
}

// prepareSiteUpdate converts Terraform site and group references into an API site update input.
func (r *lfSubPolicyResource) prepareSiteUpdate(ctx context.Context, site types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanSiteUpdateInput {
	if !utils.HasValue(site) {
		return nil
	}
	upd := r.prepareSite(ctx, site, diags)
	return &cato_models.SocketLanSiteUpdateInput{
		Group: upd.Group,
		Site:  upd.Site,
	}
}

// prepareSource converts Terraform source selectors into an API source input.
func (r *lfSubPolicyResource) prepareSource(ctx context.Context, src types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanSourceInput {
	if !utils.HasValue(src) {
		return nil
	}
	var tfSource SocketLanSource
	if utils.CheckErr(diags, src.As(ctx, &tfSource, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanSourceInput{
		FloatingSubnet:    parse.PrepareIDRefSet[cato_models.FloatingSubnetRefInput](ctx, tfSource.FloatingSubnet, diags),
		GlobalIPRange:     parse.PrepareIDRefSet[cato_models.GlobalIPRangeRefInput](ctx, tfSource.GlobalIPRange, diags),
		Group:             parse.PrepareIDRefSet[cato_models.GroupRefInput](ctx, tfSource.Group, diags),
		Host:              parse.PrepareIDRefSet[cato_models.HostRefInput](ctx, tfSource.Host, diags),
		IP:                parse.PrepareStringList[string](ctx, tfSource.IP, diags),
		IPRange:           r.prepareIPRange(ctx, tfSource.IPRange, diags),
		NetworkInterface:  parse.PrepareIDRefSet[cato_models.NetworkInterfaceRefInput](ctx, tfSource.NetworkInterface, diags),
		SiteNetworkSubnet: parse.PrepareIDRefSet[cato_models.SiteNetworkSubnetRefInput](ctx, tfSource.SiteNetworkSubnet, diags),
		Subnet:            parse.PrepareStringList[string](ctx, tfSource.Subnet, diags),
		SystemGroup:       parse.PrepareIDRefSet[cato_models.SystemGroupRefInput](ctx, tfSource.SystemGroup, diags),
		Vlan:              parse.PrepareInt64List[scalars.Vlan](ctx, tfSource.Vlan, diags),
	}
}

// prepareSourceUpdate converts Terraform source selectors into an API source update input.
func (r *lfSubPolicyResource) prepareSourceUpdate(ctx context.Context, src types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanSourceUpdateInput {
	if !utils.HasValue(src) {
		return nil
	}
	upd := r.prepareSource(ctx, src, diags)
	return &cato_models.SocketLanSourceUpdateInput{
		FloatingSubnet:    upd.FloatingSubnet,
		GlobalIPRange:     upd.GlobalIPRange,
		Group:             upd.Group,
		Host:              upd.Host,
		IP:                upd.IP,
		IPRange:           upd.IPRange,
		NetworkInterface:  upd.NetworkInterface,
		SiteNetworkSubnet: upd.SiteNetworkSubnet,
		Subnet:            upd.Subnet,
		SystemGroup:       upd.SystemGroup,
		Vlan:              upd.Vlan,
	}
}

// prepareDestination converts Terraform destination selectors into an API destination input.
func (r *lfSubPolicyResource) prepareDestination(ctx context.Context, dest types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanDestinationInput {
	if !utils.HasValue(dest) {
		return nil
	}
	var tfDestination SocketLanDestination
	if utils.CheckErr(diags, dest.As(ctx, &tfDestination, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanDestinationInput{
		FloatingSubnet:    parse.PrepareIDRefSet[cato_models.FloatingSubnetRefInput](ctx, tfDestination.FloatingSubnet, diags),
		GlobalIPRange:     parse.PrepareIDRefSet[cato_models.GlobalIPRangeRefInput](ctx, tfDestination.GlobalIPRange, diags),
		Group:             parse.PrepareIDRefSet[cato_models.GroupRefInput](ctx, tfDestination.Group, diags),
		Host:              parse.PrepareIDRefSet[cato_models.HostRefInput](ctx, tfDestination.Host, diags),
		IP:                parse.PrepareStringList[string](ctx, tfDestination.IP, diags),
		IPRange:           r.prepareIPRange(ctx, tfDestination.IPRange, diags),
		NetworkInterface:  parse.PrepareIDRefSet[cato_models.NetworkInterfaceRefInput](ctx, tfDestination.NetworkInterface, diags),
		SiteNetworkSubnet: parse.PrepareIDRefSet[cato_models.SiteNetworkSubnetRefInput](ctx, tfDestination.SiteNetworkSubnet, diags),
		Subnet:            parse.PrepareStringList[string](ctx, tfDestination.Subnet, diags),
		SystemGroup:       parse.PrepareIDRefSet[cato_models.SystemGroupRefInput](ctx, tfDestination.SystemGroup, diags),
		Vlan:              parse.PrepareInt64List[scalars.Vlan](ctx, tfDestination.Vlan, diags),
	}
}

// prepareDestinationUpdate converts Terraform destination selectors into an API destination update input.
func (r *lfSubPolicyResource) prepareDestinationUpdate(ctx context.Context, dest types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanDestinationUpdateInput {
	if !utils.HasValue(dest) {
		return nil
	}
	upd := r.prepareDestination(ctx, dest, diags)
	return &cato_models.SocketLanDestinationUpdateInput{
		FloatingSubnet:    upd.FloatingSubnet,
		GlobalIPRange:     upd.GlobalIPRange,
		Group:             upd.Group,
		Host:              upd.Host,
		IP:                upd.IP,
		IPRange:           upd.IPRange,
		NetworkInterface:  upd.NetworkInterface,
		SiteNetworkSubnet: upd.SiteNetworkSubnet,
		Subnet:            upd.Subnet,
		SystemGroup:       upd.SystemGroup,
		Vlan:              upd.Vlan,
	}
}

// prepareIPRange converts Terraform IP ranges into API address range inputs.
func (r *lfSubPolicyResource) prepareIPRange(ctx context.Context, ipRange types.List, diags *diag.Diagnostics,
) []*cato_models.IPAddressRangeInput {
	if !utils.HasValue(ipRange) {
		return nil
	}
	var tfFromTo []FromTo
	if utils.CheckErr(diags, ipRange.ElementsAs(ctx, &tfFromTo, false)) {
		return nil
	}

	out := make([]*cato_models.IPAddressRangeInput, 0, len(tfFromTo))
	for _, o := range tfFromTo {
		out = append(out, &cato_models.IPAddressRangeInput{
			From: o.From.ValueString(),
			To:   o.To.ValueString(),
		})
	}
	return out
}

// prepareNat converts Terraform NAT settings into an API input with disabled dynamic PAT defaults.
func (r *lfSubPolicyResource) prepareNat(ctx context.Context, nat types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanNatSettingsInput {
	if !utils.HasValue(nat) {
		return &cato_models.SocketLanNatSettingsInput{
			Enabled: false,
			NatType: cato_models.SocketLanNatTypeDynamicPat,
		}
	}
	var tfNat PolicyNatSettings
	if utils.CheckErr(diags, nat.As(ctx, &tfNat, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanNatSettingsInput{
		Enabled: tfNat.Enabled.ValueBool(),
		NatType: cato_models.SocketLanNatType(tfNat.NatType.ValueString()),
	}
}

// prepareNatUpdate converts Terraform NAT settings into an API update input.
func (r *lfSubPolicyResource) prepareNatUpdate(ctx context.Context, nat types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanNatSettingsUpdateInput {
	upd := r.prepareNat(ctx, nat, diags)
	return &cato_models.SocketLanNatSettingsUpdateInput{
		Enabled: &upd.Enabled,
		NatType: &upd.NatType,
	}
}

// prepareService converts Terraform custom and simple services into an API service input.
func (r *lfSubPolicyResource) prepareService(ctx context.Context, svc types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanServiceInput {
	if !utils.HasValue(svc) {
		return &cato_models.SocketLanServiceInput{}
	}
	var tfService PolicyService
	if utils.CheckErr(diags, svc.As(ctx, &tfService, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanServiceInput{
		Custom: r.prepareCustomService(ctx, tfService.Custom, diags),
		Simple: r.prepareSimpleService(ctx, tfService.Simple, diags),
	}
}

// prepareServiceUpdate converts Terraform services into an API service update input.
func (r *lfSubPolicyResource) prepareServiceUpdate(ctx context.Context, svc types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanServiceUpdateInput {
	upd := r.prepareService(ctx, svc, diags)
	return &cato_models.SocketLanServiceUpdateInput{
		Custom: upd.Custom,
		Simple: upd.Simple,
	}
}

// prepareSimpleService converts Terraform simple service names into API inputs.
func (r *lfSubPolicyResource) prepareSimpleService(ctx context.Context, svc types.Set, diags *diag.Diagnostics,
) []*cato_models.SimpleServiceInput {
	if !utils.HasValue(svc) {
		return nil
	}
	var tfSimpleServices []SimpleService
	if utils.CheckErr(diags, svc.ElementsAs(ctx, &tfSimpleServices, false)) {
		return nil
	}
	out := make([]*cato_models.SimpleServiceInput, 0, len(tfSimpleServices))
	for _, s := range tfSimpleServices {
		svcInput := cato_models.SimpleServiceInput{
			Name: cato_models.SimpleServiceType(s.Name.ValueString()),
		}
		out = append(out, &svcInput)
	}
	return out
}

// prepareCustomService converts Terraform custom service definitions into API inputs.
func (r *lfSubPolicyResource) prepareCustomService(ctx context.Context, svc types.List, diags *diag.Diagnostics,
) []*cato_models.CustomServiceInput {
	if !utils.HasValue(svc) {
		return nil
	}
	var tfCustServices []PolicyCustomService
	if utils.CheckErr(diags, svc.ElementsAs(ctx, &tfCustServices, false)) {
		return nil
	}
	out := make([]*cato_models.CustomServiceInput, 0, len(tfCustServices))
	for _, s := range tfCustServices {
		svcInput := cato_models.CustomServiceInput{
			Port:      parse.PrepareStringList[scalars.Port](ctx, s.Port, diags),
			PortRange: r.preparePortRange(ctx, s.PortRange, diags),
			Protocol:  cato_models.IPProtocol(s.Protocol.ValueString()),
		}
		out = append(out, &svcInput)
	}
	return out
}

// preparePortRange converts a Terraform port range into an API port range input.
func (r *lfSubPolicyResource) preparePortRange(ctx context.Context, portRange types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if !utils.HasValue(portRange) {
		return nil
	}
	var tfFromTo FromTo
	if utils.CheckErr(diags, portRange.As(ctx, &tfFromTo, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.PortRangeInput{
		From: scalars.Port(tfFromTo.From.ValueString()),
		To:   scalars.Port(tfFromTo.To.ValueString()),
	}
}

// hydrateLfSubPolicy fetches the current state of a lan firewall sub-policy from the API
func (r *lfSubPolicyResource) hydrateLfSubPolicy(ctx context.Context, subPolicyID string,
	planOrState *LanFirewallSubPolicy, diags *diag.Diagnostics,
) (newState *LanFirewallSubPolicy, notFound bool) {
	// Call Cato API to get the policy
	result, err := r.client.catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("failed to hydrate sub-policy", err.Error())
		return nil, false
	}

	var state *LanFirewallSubPolicy

	// Map API response to LanFirewallSubPolicy
	policy := result.GetPolicy().GetSocketLan().GetPolicy()
	for _, polRule := range policy.Rules {
		if polRule.GetSubPolicy().GetID() != subPolicyID {
			continue
		}
		if rType := polRule.GetRuleType(); rType == nil || *rType != cato_models.PolicyRuleTypeEnumSubPolicyScope {
			continue
		}
		apiRule := polRule.Rule
		state = &LanFirewallSubPolicy{
			ID:          types.StringValue(subPolicyID),
			Name:        types.StringValue(apiRule.Name),
			Description: types.StringValue(apiRule.Description),
			At:          planOrState.At,
			ScopeRuleID: types.StringValue(apiRule.ID),
			Scope:       r.parseRuleScope(ctx, apiRule, diags),
		}
		break
	}

	if state == nil {
		return nil, true
	}
	if diags.HasError() {
		return nil, false
	}

	return state, false
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

// parseNat converts API NAT settings into a Terraform object.
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

// parseService converts API custom and simple services into a Terraform object.
func (r *lfSubPolicyResource) parseService(ctx context.Context,
	svc cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service, diags *diag.Diagnostics,
) types.Object {
	if len(svc.Custom) == 0 && len(svc.Simple) == 0 {
		return types.ObjectNull(PolicyServiceTypes)
	}
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

// parseSimpleService converts API simple services into a Terraform set.
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

// parseCustomService converts API custom services into a Terraform list.
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

// parseSite converts API site and group references into a Terraform object.
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

// publish calls the API to publish the draft policy revision
func (r *lfSubPolicyResource) publish(ctx context.Context, diags *diag.Diagnostics) {
	const summary = "failed to publish LAN firewall policy"
	const notFound = "PolicyRevisionNotFound"
	result, err := r.client.catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, nil, r.client.AccountId)
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	apiErrors := result.GetPolicy().GetSocketLan().GetPublishPolicyRevision().GetErrors()
	if len(apiErrors) > 0 {
		for _, e := range apiErrors {
			if code := e.GetErrorCode(); code != nil && *code == notFound {
				continue
			}
			if msg := e.GetErrorMessage(); msg != nil {
				diags.AddError(summary, *msg)
			} else {
				diags.AddError(summary, "publishing failed")
			}
		}
		return
	}
}
