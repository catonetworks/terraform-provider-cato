package network

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

type createStaticHostAdapterFunc func(context.Context, string, string, entities.StaticHost) (string, error)

func (f createStaticHostAdapterFunc) CreateStaticHost(ctx context.Context, accountID, siteID string, host entities.StaticHost) (string, error) {
	return f(ctx, accountID, siteID, host)
}

func TestCreateStaticHostUsesEntityAndMakesOneCall(t *testing.T) {
	t.Parallel()
	mac := "00:11:22:33:44:55"
	command := CreateStaticHostCommand{AccountID: "account-123", SiteID: "site-123",
		Host: entities.StaticHost{Name: "printer", IP: "192.0.2.10", MacAddress: &mac}}
	calls := 0
	useCase, err := NewCreateStaticHostUseCase(createStaticHostAdapterFunc(func(_ context.Context, accountID, siteID string, host entities.StaticHost) (string, error) {
		calls++
		require.Equal(t, command.AccountID, accountID)
		require.Equal(t, command.SiteID, siteID)
		require.Equal(t, command.Host, host)
		return "host-123", nil
	}))
	require.NoError(t, err)
	id, err := useCase.Execute(context.Background(), command)
	require.NoError(t, err)
	require.Equal(t, "host-123", id)
	require.Equal(t, 1, calls)
}

func TestCreateStaticHostValidationAndNoMutationRetry(t *testing.T) {
	t.Parallel()
	valid := CreateStaticHostCommand{AccountID: "account-123", SiteID: "site-123",
		Host: entities.StaticHost{Name: "printer", IP: "192.0.2.10"}}
	calls := 0
	useCase, err := NewCreateStaticHostUseCase(createStaticHostAdapterFunc(func(context.Context, string, string, entities.StaticHost) (string, error) {
		calls++
		return "", catoAPIErrorStub{kind: ErrUnavailable}
	}))
	require.NoError(t, err)
	for _, command := range []CreateStaticHostCommand{
		{}, {AccountID: valid.AccountID, Host: valid.Host},
		{AccountID: valid.AccountID, SiteID: valid.SiteID, Host: entities.StaticHost{Name: "printer", IP: "bad-ip"}},
	} {
		_, err := useCase.Execute(context.Background(), command)
		var invalid *ValidationError
		require.ErrorAs(t, err, &invalid)
	}
	require.Zero(t, calls)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = useCase.Execute(ctx, valid)
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, calls)
	_, err = useCase.Execute(context.Background(), valid)
	require.True(t, errors.Is(err, ErrUnavailable))
	var failure *UseCaseError
	require.ErrorAs(t, err, &failure)
	require.NotContains(t, err.Error(), "sensitive")
	require.Equal(t, 1, calls)
}
