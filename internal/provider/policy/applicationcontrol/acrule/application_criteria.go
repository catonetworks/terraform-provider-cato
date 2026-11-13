package acrule

// Application Control application_type + application_criteria surfaces.
//
// The APC GraphQL input requires rule.applicationRule.application to select
// EXACTLY ONE of application / appCategory / customApp / customCategory /
// sanctionedAppsCategory / group / applicationType, and carries the
// attribute-based selection (compliance/security attributes, risk, origin
// country, AI risk) at the RULE level as applicationCriteria +
// applicationCriteriaSatisfy. The read (policy query) returns the same
// rule-level shape. Terraform mirrors it: application.application_type is a
// list of ApplicationType enums, and application_criteria /
// application_criteria_satisfy sit beside application_activity on the typed
// rule block.
//
// Attribute maps (compliance_attributes / security_attributes) hold ONLY
// non-ANY entries — "ANY" is the API default for an omitted key, and allowing
// it explicitly would make read-normalization ambiguous (plan-time validation
// rejects it).

import (
	"context"
	"fmt"
	"sort"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan/wfrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// acApplicationAttrTypes is the WAN application shape extended with
// application_type (APC-only selector, absent from the WAN FW schema).
var acApplicationAttrTypes = func() map[string]attr.Type {
	m := make(map[string]attr.Type, len(wfrule.WanApplicationAttrTypes)+1)
	for k, v := range wfrule.WanApplicationAttrTypes {
		m[k] = v
	}
	m["application_type"] = types.ListType{ElemType: types.StringType}
	return m
}()

var acApplicationObjectType = types.ObjectType{AttrTypes: acApplicationAttrTypes}

var acCriteriaRiskAttrTypes = map[string]attr.Type{
	"risk":          types.StringType,
	"risk_operator": types.StringType,
}

var acCriteriaAttrTypes = map[string]attr.Type{
	"compliance_attributes": types.MapType{ElemType: types.StringType},
	"security_attributes":   types.MapType{ElemType: types.StringType},
	"origin_country":        types.SetType{ElemType: idname.RefObjectType},
	"risk":                  types.ListType{ElemType: types.ObjectType{AttrTypes: acCriteriaRiskAttrTypes}},
	"ai_risk":               types.ListType{ElemType: types.StringType},
}

// acApplicationPlan mirrors wfrule.PolicyPolicyWanFirewallPolicyRulesRuleApplication
// plus application_type; element objects still decode via the wfrule item types.
type acApplicationPlan struct {
	Application            types.Set  `tfsdk:"application"`
	CustomApp              types.Set  `tfsdk:"custom_app"`
	AppCategory            types.Set  `tfsdk:"app_category"`
	CustomCategory         types.Set  `tfsdk:"custom_category"`
	SanctionedAppsCategory types.Set  `tfsdk:"sanctioned_apps_category"`
	Domain                 types.List `tfsdk:"domain"`
	Fqdn                   types.List `tfsdk:"fqdn"`
	IP                     types.List `tfsdk:"ip"`
	Subnet                 types.List `tfsdk:"subnet"`
	IPRange                types.List `tfsdk:"ip_range"`
	GlobalIPRange          types.Set  `tfsdk:"global_ip_range"`
	ApplicationType        types.List `tfsdk:"application_type"`
}

type acCriteriaPlan struct {
	ComplianceAttributes types.Map  `tfsdk:"compliance_attributes"`
	SecurityAttributes   types.Map  `tfsdk:"security_attributes"`
	OriginCountry        types.Set  `tfsdk:"origin_country"`
	Risk                 types.List `tfsdk:"risk"`
	AiRisk               types.List `tfsdk:"ai_risk"`
}

type acCriteriaRiskPlan struct {
	Risk         types.String `tfsdk:"risk"`
	RiskOperator types.String `tfsdk:"risk_operator"`
}

// ---------------------------------------------------------------------------
// write path
// ---------------------------------------------------------------------------

// acComplianceKeys maps TF map keys (exact API field names) to setters.
func acSetComplianceAttr(
	c *cato_models.ApplicationControlComplianceAttributesInput,
	key string,
	v cato_models.ApplicationControlAttributeValue,
) bool {
	switch key {
	case "hippa":
		c.Hippa = v
	case "isae3402":
		c.Isae3402 = v
	case "iso27001":
		c.Iso27001 = v
	case "pciDss":
		c.PciDss = v
	case "soc1":
		c.Soc1 = v
	case "soc2":
		c.Soc2 = v
	case "soc3":
		c.Soc3 = v
	case "sox":
		c.Sox = v
	default:
		return false
	}
	return true
}

func acSetSecurityAttr(
	s *cato_models.ApplicationControlSecurityAttributesInput,
	key string,
	v cato_models.ApplicationControlAttributeValue,
) bool {
	switch key {
	case "auditTrail":
		s.AuditTrail = v
	case "encryptionAtRest":
		s.EncryptionAtRest = v
	case "httpSecurityHeaders":
		s.HTTPSecurityHeaders = v
	case "mfa":
		s.Mfa = v
	case "rbac":
		s.Rbac = v
	case "rememberPassword":
		s.RememberPassword = v
	case "sso":
		s.Sso = v
	case "tlsEnforcement":
		s.TLSEnforcement = v
	case "trustedCertificate":
		s.TrustedCertificate = v
	default:
		return false
	}
	return true
}

func acAttributeValue(raw string, attrPath string, diags *diag.Diagnostics) cato_models.ApplicationControlAttributeValue {
	switch raw {
	case string(cato_models.ApplicationControlAttributeValueSupported):
		return cato_models.ApplicationControlAttributeValueSupported
	case string(cato_models.ApplicationControlAttributeValueNotSupported):
		return cato_models.ApplicationControlAttributeValueNotSupported
	case string(cato_models.ApplicationControlAttributeValueAny):
		diags.AddError(attrPath,
			"\"ANY\" is the API default for an omitted attribute — remove the key instead of setting it to ANY.")
		return cato_models.ApplicationControlAttributeValueAny
	default:
		diags.AddError(attrPath, fmt.Sprintf("invalid attribute value %q (expected SUPPORTED or NOT_SUPPORTED)", raw))
		return cato_models.ApplicationControlAttributeValueAny
	}
}

// acCriteriaInputFromPlan builds the rule-level applicationCriteria input.
// A null/unknown plan object yields the all-ANY default the API requires.
//
//nolint:gocyclo
func acCriteriaInputFromPlan(
	ctx context.Context,
	o types.Object,
	attrPath string,
	diags diag.Diagnostics,
) (*cato_models.ApplicationControlCriteriaInput, diag.Diagnostics) {
	out := acEmptyApplicationCriteriaInput()
	if o.IsNull() || o.IsUnknown() {
		return out, diags
	}
	p := acCriteriaPlan{}
	diags.Append(o.As(ctx, &p, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return out, diags
	}

	if !p.ComplianceAttributes.IsNull() && !p.ComplianceAttributes.IsUnknown() {
		vals := map[string]string{}
		diags.Append(p.ComplianceAttributes.ElementsAs(ctx, &vals, false)...)
		for k, raw := range vals {
			v := acAttributeValue(raw, attrPath+".compliance_attributes."+k, &diags)
			if !acSetComplianceAttr(out.Attributes.ComplianceAttributes, k, v) {
				diags.AddError(attrPath+".compliance_attributes",
					fmt.Sprintf("unknown compliance attribute %q (valid: hippa, isae3402, iso27001, pciDss, soc1, soc2, soc3, sox)", k))
			}
		}
	}
	if !p.SecurityAttributes.IsNull() && !p.SecurityAttributes.IsUnknown() {
		vals := map[string]string{}
		diags.Append(p.SecurityAttributes.ElementsAs(ctx, &vals, false)...)
		for k, raw := range vals {
			v := acAttributeValue(raw, attrPath+".security_attributes."+k, &diags)
			if !acSetSecurityAttr(out.Attributes.SecurityAttributes, k, v) {
				diags.AddError(attrPath+".security_attributes",
					fmt.Sprintf("unknown security attribute %q (valid: auditTrail, encryptionAtRest, httpSecurityHeaders, "+
						"mfa, rbac, rememberPassword, sso, tlsEnforcement, trustedCertificate)", k))
			}
		}
	}
	if !p.OriginCountry.IsNull() && !p.OriginCountry.IsUnknown() {
		objs := make([]types.Object, 0, len(p.OriginCountry.Elements()))
		diags.Append(p.OriginCountry.ElementsAs(ctx, &objs, false)...)
		for _, co := range objs {
			var item wfrule.PolicyPolicyWanFirewallPolicyRulesRuleSourceHost
			diags.Append(co.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			ref, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.AddError(attrPath+".origin_country", err.Error())
				continue
			}
			out.OriginCountry = append(out.OriginCountry, &cato_models.CountryRefInput{
				By: cato_models.ObjectRefBy(ref.By), Input: ref.Input,
			})
		}
	}
	if !p.Risk.IsNull() && !p.Risk.IsUnknown() {
		rows := make([]acCriteriaRiskPlan, 0, len(p.Risk.Elements()))
		diags.Append(p.Risk.ElementsAs(ctx, &rows, false)...)
		for _, row := range rows {
			out.Risk = append(out.Risk, &cato_models.ApplicationControlRiskCriteriaInput{
				Risk:         scalars.ApplicationRisk(row.Risk.ValueString()),
				RiskOperator: cato_models.ApplicationControlOperator(row.RiskOperator.ValueString()),
			})
		}
	}
	if !p.AiRisk.IsNull() && !p.AiRisk.IsUnknown() {
		var vals []string
		diags.Append(p.AiRisk.ElementsAs(ctx, &vals, false)...)
		for _, v := range vals {
			out.AiRisk = append(out.AiRisk, cato_models.ApplicationControlAiRisk(v))
		}
	}
	return out, diags
}

// acCriteriaUpdateFromInput converts an add-shaped criteria input to the
// update-shaped variant (same values, pointer/omitempty fields).
func acCriteriaUpdateFromInput(in *cato_models.ApplicationControlCriteriaInput) *cato_models.ApplicationControlCriteriaUpdateInput {
	if in == nil {
		return nil
	}
	out := &cato_models.ApplicationControlCriteriaUpdateInput{
		AiRisk:        in.AiRisk,
		OriginCountry: in.OriginCountry,
		Risk:          in.Risk,
	}
	if in.Attributes != nil {
		out.Attributes = &cato_models.ApplicationControlAttributesUpdateInput{}
		if c := in.Attributes.ComplianceAttributes; c != nil {
			cc := *c
			out.Attributes.ComplianceAttributes = &cato_models.ApplicationControlComplianceAttributesUpdateInput{
				Hippa: &cc.Hippa, Isae3402: &cc.Isae3402, Iso27001: &cc.Iso27001, PciDss: &cc.PciDss,
				Soc1: &cc.Soc1, Soc2: &cc.Soc2, Soc3: &cc.Soc3, Sox: &cc.Sox,
			}
		}
		if s := in.Attributes.SecurityAttributes; s != nil {
			ss := *s
			out.Attributes.SecurityAttributes = &cato_models.ApplicationControlSecurityAttributesUpdateInput{
				AuditTrail: &ss.AuditTrail, EncryptionAtRest: &ss.EncryptionAtRest,
				HTTPSecurityHeaders: &ss.HTTPSecurityHeaders, Mfa: &ss.Mfa, Rbac: &ss.Rbac,
				RememberPassword: &ss.RememberPassword, Sso: &ss.Sso,
				TLSEnforcement: &ss.TLSEnforcement, TrustedCertificate: &ss.TrustedCertificate,
			}
		}
	}
	return out
}

// acApplicationTypesFromPlan extracts application.application_type values.
func acApplicationTypesFromPlan(ctx context.Context, l types.List, diags *diag.Diagnostics) []cato_models.ApplicationType {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var vals []string
	diags.Append(l.ElementsAs(ctx, &vals, false)...)
	out := make([]cato_models.ApplicationType, 0, len(vals))
	for _, v := range vals {
		switch cato_models.ApplicationType(v) {
		case cato_models.ApplicationTypeApplication, cato_models.ApplicationTypeCloudApplication, cato_models.ApplicationTypeService:
			out = append(out, cato_models.ApplicationType(v))
		default:
			diags.AddError("application.application_type",
				fmt.Sprintf("invalid application type %q (expected APPLICATION, CLOUD_APPLICATION or SERVICE)", v))
		}
	}
	return out
}

// acApplicationSelectsSomething reports whether the application input selects
// at least one matcher — the API requires exactly one of the selector fields
// and rejects both a null and an empty application object.
func acApplicationSelectsSomething(
	a *cato_models.ApplicationControlApplicationInput,
) bool {
	if a == nil {
		return false
	}
	return a.Application != nil || a.CustomApp != nil || a.AppCategory != nil ||
		a.CustomCategory != nil || a.SanctionedAppsCategory != nil || len(a.ApplicationType) > 0
}

// acRequireApplicationSelection emits a friendly diagnostics error instead of
// letting the API reject the mutation with an opaque GraphQL non-null error.
func acRequireApplicationSelection(
	a *cato_models.ApplicationControlApplicationInput,
	rulePath string,
	diags *diag.Diagnostics,
) {
	if acApplicationSelectsSomething(a) {
		return
	}
	diags.AddError(
		rulePath+".application selection required",
		"The Cato API requires the application block to select exactly one of: "+
			"application, custom_app, app_category, custom_category, sanctioned_apps_category or application_type. "+
			"Rules that match applications by attribute (compliance/security/risk/origin-country criteria) still "+
			"need application_type (e.g. [\"CLOUD_APPLICATION\"]) plus application_criteria.",
	)
}

// ---------------------------------------------------------------------------
// read path
// ---------------------------------------------------------------------------

// acApplicationTypeListValue renders applicationType read values ([] when none).
func acApplicationTypeListValue(appTypes []cato_models.ApplicationType, diags *diag.Diagnostics) types.List {
	vals := make([]attr.Value, 0, len(appTypes))
	for _, t := range appTypes {
		vals = append(vals, types.StringValue(string(t)))
	}
	l, d := types.ListValue(types.StringType, vals)
	diags.Append(d...)
	return l
}

// acCriteriaState is the rule-type-agnostic read shape the generated SDK
// response structs are adapted into (one adapter per rule type below).
type acCriteriaState struct {
	Compliance map[string]string // only non-ANY entries
	Security   map[string]string // only non-ANY entries
	Countries  []acNameIDState
	Risk       []acCriteriaRiskState
	AiRisk     []string
}

type acNameIDState struct {
	ID   string
	Name string
}

type acCriteriaRiskState struct {
	Risk         string
	RiskOperator string
}

func (c acCriteriaState) isTrivial() bool {
	return len(c.Compliance) == 0 && len(c.Security) == 0 &&
		len(c.Countries) == 0 && len(c.Risk) == 0 && len(c.AiRisk) == 0
}

func acNonAny(m map[string]string, key, val string) {
	if val != "" && val != string(cato_models.ApplicationControlAttributeValueAny) {
		m[key] = val
	}
}

// acCriteriaObjectValue renders criteria state; trivial (all-ANY/empty)
// criteria render as null so rules without criteria keep clean plans.
func acCriteriaObjectValue(c acCriteriaState, diags *diag.Diagnostics) types.Object {
	if c.isTrivial() {
		return types.ObjectNull(acCriteriaAttrTypes)
	}
	toMap := func(src map[string]string) types.Map {
		vals := map[string]attr.Value{}
		keys := make([]string, 0, len(src))
		for k := range src {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			vals[k] = types.StringValue(src[k])
		}
		m, d := types.MapValue(types.StringType, vals)
		diags.Append(d...)
		return m
	}
	compliance := types.MapNull(types.StringType)
	if len(c.Compliance) > 0 {
		compliance = toMap(c.Compliance)
	}
	security := types.MapNull(types.StringType)
	if len(c.Security) > 0 {
		security = toMap(c.Security)
	}

	countrySet := types.SetNull(idname.RefObjectType)
	if len(c.Countries) > 0 {
		vals := make([]attr.Value, 0, len(c.Countries))
		for _, co := range c.Countries {
			o, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(co.ID),
				"name": types.StringValue(co.Name),
			})
			diags.Append(d...)
			vals = append(vals, o)
		}
		s, d := types.SetValue(idname.RefObjectType, vals)
		diags.Append(d...)
		countrySet = s
	}

	riskList := types.ListNull(types.ObjectType{AttrTypes: acCriteriaRiskAttrTypes})
	if len(c.Risk) > 0 {
		vals := make([]attr.Value, 0, len(c.Risk))
		for _, r := range c.Risk {
			o, d := types.ObjectValue(acCriteriaRiskAttrTypes, map[string]attr.Value{
				"risk":          types.StringValue(r.Risk),
				"risk_operator": types.StringValue(r.RiskOperator),
			})
			diags.Append(d...)
			vals = append(vals, o)
		}
		l, d := types.ListValue(types.ObjectType{AttrTypes: acCriteriaRiskAttrTypes}, vals)
		diags.Append(d...)
		riskList = l
	}

	aiList := types.ListNull(types.StringType)
	if len(c.AiRisk) > 0 {
		vals := make([]attr.Value, 0, len(c.AiRisk))
		for _, v := range c.AiRisk {
			vals = append(vals, types.StringValue(v))
		}
		l, d := types.ListValue(types.StringType, vals)
		diags.Append(d...)
		aiList = l
	}

	o, d := types.ObjectValue(acCriteriaAttrTypes, map[string]attr.Value{
		"compliance_attributes": compliance,
		"security_attributes":   security,
		"origin_country":        countrySet,
		"risk":                  riskList,
		"ai_risk":               aiList,
	})
	diags.Append(d...)
	return o
}

// acComplianceStateMap/acSecurityStateMap collect non-ANY attribute values.
// The generated response structs differ per rule type but all expose the same
// enum-typed fields, so adapters pass plain values here.
func acComplianceStateMap(
	hippa, isae3402, iso27001, pciDss, soc1, soc2, soc3, sox cato_models.ApplicationControlAttributeValue,
) map[string]string {
	m := map[string]string{}
	acNonAny(m, "hippa", string(hippa))
	acNonAny(m, "isae3402", string(isae3402))
	acNonAny(m, "iso27001", string(iso27001))
	acNonAny(m, "pciDss", string(pciDss))
	acNonAny(m, "soc1", string(soc1))
	acNonAny(m, "soc2", string(soc2))
	acNonAny(m, "soc3", string(soc3))
	acNonAny(m, "sox", string(sox))
	return m
}

func acSecurityStateMap(
	auditTrail, encryptionAtRest, httpSecurityHeaders, mfa, rbac, rememberPassword,
	sso, tlsEnforcement, trustedCertificate cato_models.ApplicationControlAttributeValue,
) map[string]string {
	m := map[string]string{}
	acNonAny(m, "auditTrail", string(auditTrail))
	acNonAny(m, "encryptionAtRest", string(encryptionAtRest))
	acNonAny(m, "httpSecurityHeaders", string(httpSecurityHeaders))
	acNonAny(m, "mfa", string(mfa))
	acNonAny(m, "rbac", string(rbac))
	acNonAny(m, "rememberPassword", string(rememberPassword))
	acNonAny(m, "sso", string(sso))
	acNonAny(m, "tlsEnforcement", string(tlsEnforcement))
	acNonAny(m, "trustedCertificate", string(trustedCertificate))
	return m
}

func acAiRiskStrings(in []cato_models.ApplicationControlAiRisk) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, string(v))
	}
	return out
}

// acSatisfyStateValue normalizes the rule-level criteria satisfy for state.
func acSatisfyStateValue(s cato_models.ApplicationControlSatisfy) types.String {
	if strings.TrimSpace(string(s)) == "" {
		return types.StringValue(string(cato_models.ApplicationControlSatisfyAll))
	}
	return types.StringValue(string(s))
}

// ---------------------------------------------------------------------------
// per-rule-type read adapters — the generated response structs are identical
// in shape but distinct Go types, so each adapter flattens into acCriteriaState.
// ---------------------------------------------------------------------------

func acCriteriaObjectFromApplicationRule(
	c *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_ApplicationRule_ApplicationCriteria,
	diags *diag.Diagnostics,
) types.Object {
	if c == nil {
		return types.ObjectNull(acCriteriaAttrTypes)
	}
	st := acCriteriaState{
		Compliance: acComplianceStateMap(
			c.Attributes.ComplianceAttributes.Hippa, c.Attributes.ComplianceAttributes.Isae3402,
			c.Attributes.ComplianceAttributes.Iso27001, c.Attributes.ComplianceAttributes.PciDss,
			c.Attributes.ComplianceAttributes.Soc1, c.Attributes.ComplianceAttributes.Soc2,
			c.Attributes.ComplianceAttributes.Soc3, c.Attributes.ComplianceAttributes.Sox,
		),
		Security: acSecurityStateMap(
			c.Attributes.SecurityAttributes.AuditTrail, c.Attributes.SecurityAttributes.EncryptionAtRest,
			c.Attributes.SecurityAttributes.HTTPSecurityHeaders, c.Attributes.SecurityAttributes.Mfa,
			c.Attributes.SecurityAttributes.Rbac, c.Attributes.SecurityAttributes.RememberPassword,
			c.Attributes.SecurityAttributes.Sso, c.Attributes.SecurityAttributes.TLSEnforcement,
			c.Attributes.SecurityAttributes.TrustedCertificate,
		),
		AiRisk: acAiRiskStrings(c.AiRisk),
	}
	for _, co := range c.OriginCountry {
		if co != nil {
			st.Countries = append(st.Countries, acNameIDState{ID: co.ID, Name: co.Name})
		}
	}
	for _, r := range c.Risk {
		if r != nil {
			st.Risk = append(st.Risk, acCriteriaRiskState{Risk: r.Risk.String(), RiskOperator: string(r.RiskOperator)})
		}
	}
	return acCriteriaObjectValue(st, diags)
}

func acCriteriaObjectFromDataRule(
	c *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_DataRule_ApplicationCriteria,
	diags *diag.Diagnostics,
) types.Object {
	if c == nil {
		return types.ObjectNull(acCriteriaAttrTypes)
	}
	st := acCriteriaState{
		Compliance: acComplianceStateMap(
			c.Attributes.ComplianceAttributes.Hippa, c.Attributes.ComplianceAttributes.Isae3402,
			c.Attributes.ComplianceAttributes.Iso27001, c.Attributes.ComplianceAttributes.PciDss,
			c.Attributes.ComplianceAttributes.Soc1, c.Attributes.ComplianceAttributes.Soc2,
			c.Attributes.ComplianceAttributes.Soc3, c.Attributes.ComplianceAttributes.Sox,
		),
		Security: acSecurityStateMap(
			c.Attributes.SecurityAttributes.AuditTrail, c.Attributes.SecurityAttributes.EncryptionAtRest,
			c.Attributes.SecurityAttributes.HTTPSecurityHeaders, c.Attributes.SecurityAttributes.Mfa,
			c.Attributes.SecurityAttributes.Rbac, c.Attributes.SecurityAttributes.RememberPassword,
			c.Attributes.SecurityAttributes.Sso, c.Attributes.SecurityAttributes.TLSEnforcement,
			c.Attributes.SecurityAttributes.TrustedCertificate,
		),
		AiRisk: acAiRiskStrings(c.AiRisk),
	}
	for _, co := range c.OriginCountry {
		if co != nil {
			st.Countries = append(st.Countries, acNameIDState{ID: co.ID, Name: co.Name})
		}
	}
	for _, r := range c.Risk {
		if r != nil {
			st.Risk = append(st.Risk, acCriteriaRiskState{Risk: r.Risk.String(), RiskOperator: string(r.RiskOperator)})
		}
	}
	return acCriteriaObjectValue(st, diags)
}

func acCriteriaObjectFromFileRule(
	c *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule_FileRule_ApplicationCriteria,
	diags *diag.Diagnostics,
) types.Object {
	if c == nil {
		return types.ObjectNull(acCriteriaAttrTypes)
	}
	st := acCriteriaState{
		Compliance: acComplianceStateMap(
			c.Attributes.ComplianceAttributes.Hippa, c.Attributes.ComplianceAttributes.Isae3402,
			c.Attributes.ComplianceAttributes.Iso27001, c.Attributes.ComplianceAttributes.PciDss,
			c.Attributes.ComplianceAttributes.Soc1, c.Attributes.ComplianceAttributes.Soc2,
			c.Attributes.ComplianceAttributes.Soc3, c.Attributes.ComplianceAttributes.Sox,
		),
		Security: acSecurityStateMap(
			c.Attributes.SecurityAttributes.AuditTrail, c.Attributes.SecurityAttributes.EncryptionAtRest,
			c.Attributes.SecurityAttributes.HTTPSecurityHeaders, c.Attributes.SecurityAttributes.Mfa,
			c.Attributes.SecurityAttributes.Rbac, c.Attributes.SecurityAttributes.RememberPassword,
			c.Attributes.SecurityAttributes.Sso, c.Attributes.SecurityAttributes.TLSEnforcement,
			c.Attributes.SecurityAttributes.TrustedCertificate,
		),
		AiRisk: acAiRiskStrings(c.AiRisk),
	}
	for _, co := range c.OriginCountry {
		if co != nil {
			st.Countries = append(st.Countries, acNameIDState{ID: co.ID, Name: co.Name})
		}
	}
	for _, r := range c.Risk {
		if r != nil {
			st.Risk = append(st.Risk, acCriteriaRiskState{Risk: r.Risk.String(), RiskOperator: string(r.RiskOperator)})
		}
	}
	return acCriteriaObjectValue(st, diags)
}

// acNameIDSetValue renders a name/id ref set for state (null when empty, so
// configs that omit the attribute stay clean).
func acNameIDSetValue(items []acNameIDState, diags *diag.Diagnostics) types.Set {
	if len(items) == 0 {
		return types.SetNull(idname.RefObjectType)
	}
	vals := make([]attr.Value, 0, len(items))
	for _, it := range items {
		o, d := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(it.ID),
			"name": types.StringValue(it.Name),
		})
		diags.Append(d...)
		vals = append(vals, o)
	}
	s, d := types.SetValue(idname.RefObjectType, vals)
	diags.Append(d...)
	return s
}

// ---------------------------------------------------------------------------
// plan modifiers for the optional ref attributes (activity field/value_set,
// file-attribute content-type sets, criteria origin_country). The shared
// IDName modifiers resolve configured refs (id-configured → name unknown and
// vice versa) but leave a NULL-config attribute unknown, which plans as
// perpetual "(known after apply)" churn — these wrappers pin null config to a
// null plan first, then delegate.
// ---------------------------------------------------------------------------

type acRefObjectModifier struct{ inner planmodifier.Object }

func acRefObjectPlanModifier() planmodifier.Object {
	return acRefObjectModifier{inner: planmodifiers.IDNamePlanModifier()}
}

func (m acRefObjectModifier) Description(_ context.Context) string {
	return "null config plans null; configured id/name refs resolve via the IDName modifier"
}
func (m acRefObjectModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m acRefObjectModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if !req.ConfigValue.IsUnknown() && req.ConfigValue.IsNull() && !req.Plan.Raw.IsNull() {
		resp.PlanValue = types.ObjectNull(idname.RefAttrTypes)
		return
	}
	m.inner.PlanModifyObject(ctx, req, resp)
}

type acRefSetModifier struct{ inner planmodifier.Set }

func acRefSetPlanModifier() planmodifier.Set {
	return acRefSetModifier{inner: planmodifiers.IDNameSetPlanModifier()}
}

func (m acRefSetModifier) Description(_ context.Context) string {
	return "null config plans null; configured id/name ref sets resolve via the IDNameSet modifier"
}
func (m acRefSetModifier) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }

func (m acRefSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if !req.ConfigValue.IsUnknown() && req.ConfigValue.IsNull() && !req.Plan.Raw.IsNull() {
		resp.PlanValue = types.SetNull(idname.RefObjectType)
		return
	}
	m.inner.PlanModifySet(ctx, req, resp)
}
