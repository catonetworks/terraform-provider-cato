package alwaysonsection

import (
	"context"
	"errors"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

// ResourceClient contains the SDK operations used by the Always On section resource.
type ResourceClient interface {
	PolicyZtnaAlwaysOnAddSection(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyAddSectionInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnAddSection, error)
	PolicyZtnaAlwaysOnPolicy(ctx context.Context, accountID string, input *cato_models.ZtnaAlwaysOnPolicyInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnPolicy, error)
	PolicyZtnaAlwaysOnUpdateSection(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyUpdateSectionInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnUpdateSection, error)
	PolicyZtnaAlwaysOnMoveSection(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnMoveSection, error)
	PolicyZtnaAlwaysOnRemoveSection(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyRemoveSectionInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnRemoveSection, error)
}

func (r *alwaysOnSectionResource) configuredClient() ResourceClient {
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{resource: r}
}

type sdkClient struct{ resource *alwaysOnSectionResource }

var _ ResourceClient = (*sdkClient)(nil)

func (c *sdkClient) PolicyZtnaAlwaysOnAddSection(ctx context.Context, accountID string,
	policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyAddSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnAddSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnAddSection(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnPolicy(ctx context.Context, accountID string,
	input *cato_models.ZtnaAlwaysOnPolicyInput, interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnPolicy, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnPolicy(ctx, accountID, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnUpdateSection(ctx context.Context, accountID string,
	policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyUpdateSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnUpdateSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnUpdateSection(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnMoveSection(ctx context.Context, accountID string,
	policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyMoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnMoveSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnMoveSection(ctx, accountID, policyInput, input, interceptors...)
}

func (c *sdkClient) PolicyZtnaAlwaysOnRemoveSection(ctx context.Context, accountID string,
	policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyRemoveSectionInput,
	interceptors ...clientv2.RequestInterceptor,
) (*cato.PolicyZtnaAlwaysOnRemoveSection, error) {
	return c.resource.client.Catov2.PolicyZtnaAlwaysOnRemoveSection(ctx, accountID, policyInput, input, interceptors...)
}

func findSection(result *cato.PolicyZtnaAlwaysOnPolicy, id string,
) (*cato.PolicyZtnaAlwaysOnPolicy_Policy_ZtnaAlwaysOn_Policy_Sections_Section,
	*cato_models.PolicySectionPositionInput, error,
) {
	policy := result.GetPolicy().GetZtnaAlwaysOn()
	if policy == nil || policy.GetPolicy().GetSections() == nil {
		return nil, nil, errors.New("API did not return an Always On policy with a section list")
	}
	sections := policy.GetPolicy().GetSections()
	index := -1
	seen := make(map[string]bool, len(sections))
	for i, entry := range sections {
		sectionID := entry.GetSection().GetID()
		if sectionID == "" {
			return nil, nil, errors.New("API returned an Always On section without an ID")
		}
		if seen[sectionID] {
			return nil, nil, errors.New("API returned duplicate Always On section IDs")
		}
		seen[sectionID] = true
		if sectionID == id {
			index = i
		}
	}
	if index < 0 {
		return nil, nil, nil
	}
	position := &cato_models.PolicySectionPositionInput{Position: cato_models.PolicySectionPositionEnumLastInPolicy}
	if index+1 < len(sections) {
		nextID := sections[index+1].GetSection().GetID()
		position.Position, position.Ref = cato_models.PolicySectionPositionEnumBeforeSection, &nextID
	}
	return sections[index].GetSection(), position, nil
}

func readSectionByID(ctx context.Context, client ResourceClient, accountID, id string) (sectionInfo, error) {
	result, err := client.PolicyZtnaAlwaysOnPolicy(ctx, accountID, nil)
	if err != nil {
		return nil, err
	}
	section, _, err := findSection(result, id)
	if err != nil {
		return nil, err
	}
	if section == nil {
		return nil, errors.New("API did not return the updated Always On section")
	}
	return section, nil
}
