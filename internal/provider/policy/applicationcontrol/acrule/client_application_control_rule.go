package acrule

import (
	"context"
	"fmt"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
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

// ResourceClient contains only the SDK operations used by this resource.
type ResourceClient interface {
	ApplicationControlPolicy(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.ApplicationControlPolicy, error)
	PolicyApplicationControlAddRule(
		ctx context.Context,
		input cato_models.ApplicationControlAddRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlAddRule, error)
	PolicyApplicationControlMoveRule(
		ctx context.Context,
		input cato_models.PolicyMoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlMoveRule, error)
	PolicyApplicationControlRemoveRule(
		ctx context.Context,
		input cato_models.ApplicationControlRemoveRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlRemoveRule, error)
	PolicyApplicationControlUpdateRule(
		ctx context.Context,
		input cato_models.ApplicationControlUpdateRuleInput,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlUpdateRule, error)
	PolicyApplicationControlPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error)
}

func (r *applicationControlRuleResource) resourceAPI() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	return &sdkClient{resource: r}
}

func (c *sdkClient) ApplicationControlPolicy(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.ApplicationControlPolicy, error) {
	return c.resource.client.Catov2.ApplicationControlPolicy(ctx, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlAddRule(
	ctx context.Context,
	input cato_models.ApplicationControlAddRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlAddRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlAddRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlMoveRule(
	ctx context.Context,
	input cato_models.PolicyMoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlMoveRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlMoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlRemoveRule(
	ctx context.Context,
	input cato_models.ApplicationControlRemoveRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlRemoveRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlRemoveRule(ctx, input, accountID, interceptors...)
}

func (c *sdkClient) PolicyApplicationControlUpdateRule(
	ctx context.Context,
	input cato_models.ApplicationControlUpdateRuleInput,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlUpdateRule, error) {
	return c.resource.client.Catov2.PolicyApplicationControlUpdateRule(ctx, input, accountID, interceptors...)
}

// sdkClient forwards operations to the configured SDK client.
type sdkClient struct {
	resource *applicationControlRuleResource
}

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyApplicationControlPublishPolicyRevision(
	ctx context.Context,
	accountID string,
	interceptors ...clientv2.RequestInterceptor,
) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error) {
	return c.resource.client.Catov2.PolicyApplicationControlPublishPolicyRevision(ctx, accountID, interceptors...)
}
