package alwaysonpolicy

import (
	"context"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestPolicyReadMetadataAndClearMissingDraft(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := mocks.NewAlwaysOnPolicyClient(t)
	withDraft := decodeResponse[cato.PolicyZtnaAlwaysOnPolicy](t, `{"policy":{"ztnaAlwaysOn":{"policy":{
		"enabled":true,"audit":{"publishedTime":"2026-11-03T10:00:00Z","publishedBy":"admin"},
		"revision":{"id":"draft-1","name":"Draft"}}}}}`)
	client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", mock.Anything).Return(withDraft, nil).Once()
	client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", mock.Anything).Return(policyResponse(t, true), nil).Once()
	r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	state := policyState(t, types.BoolValue(true))
	for i := range 2 {
		resp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		var actual PolicyModel
		require.False(t, resp.State.Get(ctx, &actual).HasError())
		if i == 0 {
			require.Equal(t, types.StringValue("admin"), actual.PublishedBy)
			require.Equal(t, types.StringValue("2026-11-03T10:00:00Z"), actual.PublishedTime)
			require.Equal(t, types.StringValue("draft-1"), actual.RevisionID)
			require.Equal(t, types.StringValue("Draft"), actual.RevisionName)
		} else {
			require.True(t, actual.PublishedBy.IsNull())
			require.True(t, actual.PublishedTime.IsNull())
			require.True(t, actual.RevisionID.IsNull())
			require.True(t, actual.RevisionName.IsNull())
		}
		state = tfsdk.State{Schema: resp.State.Schema, Raw: resp.State.Raw}
	}
}

func TestPolicyPublicationRetryHonorsCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := mocks.NewAlwaysOnPolicyClient(t)
	client.On("PolicyZtnaAlwaysOnPublishPolicyRevision", mock.Anything, "123", mock.Anything, mock.Anything).
		Return(nil, errors.New("reorderPolicyBlockedByActiveSessions")).Run(func(mock.Arguments) { cancel() }).Once()
	r := &alwaysOnPolicyResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	resp := resource.DeleteResponse{State: policyState(t, types.BoolValue(true))}
	r.Delete(ctx, resource.DeleteRequest{}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "context canceled")
}
