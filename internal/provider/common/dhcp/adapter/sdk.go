package adapter

import (
	"context"
	"fmt"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"

	apperr "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
)

type Client interface {
	EntityLookupMinimal(
		context.Context,
		string,
		models.EntityType,
		*int64,
		*int64,
		*models.EntityInput,
		[]*models.SortInput,
		[]*models.LookupFilterInput,
		...clientv2.RequestInterceptor,
	) (*cato.EntityLookup, error)
}

type Lookup struct {
	Client    Client
	AccountID string
}

func (l Lookup) groups(ctx context.Context) (*cato.EntityLookup, error) {
	result, err := l.Client.EntityLookupMinimal(
		ctx,
		l.AccountID,
		models.EntityTypeDhcpRelayGroup,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, &apperr.Error{
			Summary: "Failed to lookup DHCP relay group",
			Detail:  fmt.Sprintf("An error was encountered when looking up DHCP relay group: %v", err),
			Cause:   err,
		}
	}
	if result == nil {
		return nil, apperr.New("Failed to lookup DHCP relay group", "DHCP relay groups were not returned")
	}
	return result, nil
}
func (l Lookup) RelayID(ctx context.Context, name string) (string, error) {
	result, err := l.groups(ctx)
	if err != nil {
		return "", err
	}
	for _, item := range result.EntityLookup.Items {
		if n := item.Entity.GetName(); n != nil && *n == name {
			return item.Entity.GetID(), nil
		}
	}
	return "", apperr.New(
		"Failed to lookup DHCP relay group",
		fmt.Sprintf("DHCP relay group: '%s' not found", name),
	)
}
func (l Lookup) RelayName(ctx context.Context, id string) (*string, error) {
	result, err := l.groups(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range result.EntityLookup.Items {
		if item.Entity.GetID() == id {
			return item.Entity.GetName(), nil
		}
	}
	return nil, apperr.New("Failed to lookup DHCP relay group", fmt.Sprintf("DHCP relay group: '%s' not found", id))
}
func Input(s *application.Settings) *models.NetworkDhcpSettingsInput {
	if s == nil {
		return nil
	}
	return &models.NetworkDhcpSettingsInput{

		DhcpType:              models.DhcpType(s.DhcpType),
		IPRange:               s.IPRange,
		RelayGroupID:          s.RelayGroupID,
		DhcpMicrosegmentation: s.DhcpMicrosegmentation,
	}
}
func Snapshot(kind string, ip, relay *string, micro bool) *application.Settings {
	return &application.Settings{

		DhcpType:              kind,
		IPRange:               ip,
		RelayGroupID:          relay,
		DhcpMicrosegmentation: &micro,
	}
}
