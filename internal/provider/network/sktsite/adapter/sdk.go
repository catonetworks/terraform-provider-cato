package adapter

import (
	"context"
	"fmt"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"

	apperr "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/client"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/adapter"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/application"
)

type SocketSiteClient interface {
	dhcp.Client
	SiteAddSocketSite(context.Context, models.AddSocketSiteInput, string, ...clientv2.RequestInterceptor) (*cato.SiteAddSocketSite, error)
	SiteSocketConfiguration(
		context.Context,
		models.SiteSocketConfigurationInput,
		string,
		...clientv2.RequestInterceptor,
	) (
		*cato.SiteSocketConfiguration,
		error,
	)
	SiteGeneralDetails(context.Context, models.SiteRefInput, string, ...clientv2.RequestInterceptor) (*cato.SiteGeneralDetails, error)
	SiteUpdateSiteGeneralDetails(
		context.Context,
		string,
		models.UpdateSiteGeneralDetailsInput,
		string,
		...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateSiteGeneralDetails, error)
	SiteUpdateNetworkRange(
		context.Context,
		string,
		models.UpdateNetworkRangeInput,
		string,
		...clientv2.RequestInterceptor,
	) (
		*cato.SiteUpdateNetworkRange,
		error,
	)
	SiteUpdateSocketInterface(
		context.Context,
		string,
		models.SocketInterfaceIDEnum,
		models.UpdateSocketInterfaceInput,
		string,
		...clientv2.RequestInterceptor,
	) (*cato.SiteUpdateSocketInterface, error)
	SiteExchangeSocketPorts(
		context.Context,
		string,
		models.ExchangeSocketPortsInput,
		...clientv2.RequestInterceptor,
	) (
		*cato.SiteExchangeSocketPorts,
		error,
	)
	NetworkRangeList(context.Context, string, models.NetworkRangeListInput, ...clientv2.RequestInterceptor) (*cato.NetworkRangeList, error)
	EntityLookup(
		context.Context,
		string,
		models.EntityType,
		*int64,
		*int64,
		*models.EntityInput,
		*string,
		[]string,
		[]*models.SortInput,
		[]*models.LookupFilterInput,
		[]string,
		...clientv2.RequestInterceptor,
	) (*cato.EntityLookup, error)
	SiteRemoveSite(context.Context, string, string, ...clientv2.RequestInterceptor) (*cato.SiteRemoveSite, error)
}
type SDK struct {
	Client    SocketSiteClient
	AccountID string
}

var _ application.SocketSitePort = SDK{}

func ResolveClient(injected SocketSiteClient, data *client.CatoClientData) SocketSiteClient {
	if injected != nil {
		return injected
	}
	if data == nil {
		return nil
	}
	return data.Catov2
}
func (a SDK) RelayID(ctx context.Context, name string) (string, error) {
	return (dhcp.Lookup{
		Client:    a.Client,
		AccountID: a.AccountID,
	}).RelayID(ctx, name)
}
func (a SDK) RelayName(ctx context.Context, id string) (*string, error) {
	return (dhcp.Lookup{
		Client:    a.Client,
		AccountID: a.AccountID,
	}).RelayName(ctx, id)
}
func (a SDK) Add(ctx context.Context, in application.AddInput) (string, error) {
	input := models.AddSocketSiteInput{

		ConnectionType:     models.SiteConnectionTypeEnum(in.ConnectionType),
		Name:               in.Name,
		NativeNetworkRange: in.NativeNetworkRange,
		SiteType:           models.SiteType(in.SiteType),
		Description:        in.Description,
		TranslatedSubnet:   in.TranslatedSubnet,
		Vlan:               (*scalars.Vlan)(in.Vlan),
	}
	if l := in.SiteLocation; l != nil {
		input.SiteLocation = &models.AddSiteLocationInput{

			Address:     l.Address,
			City:        l.City,
			CountryCode: l.CountryCode,
			StateCode:   l.StateCode,
			Timezone:    l.Timezone,
		}
	}
	result, err := a.Client.SiteAddSocketSite(ctx, input, a.AccountID)
	if err != nil {
		return "", apperr.Wrap("Catov2 API SiteAddSocketSite error", err)
	}
	if result == nil || result.Site.AddSocketSite == nil || result.Site.AddSocketSite.GetSiteID() == "" {
		return "", apperr.Wrap("Catov2 API SiteAddSocketSite error", fmt.Errorf("empty site ID returned from API"))
	}
	return result.Site.AddSocketSite.GetSiteID(), nil
}
func (a SDK) UpdateGeneral(ctx context.Context, id string, in application.GeneralInput) error {
	input := models.UpdateSiteGeneralDetailsInput{

		Name:        in.Name,
		Description: in.Description,
		SiteType:    (*models.SiteType)(in.SiteType),
	}
	if l := in.SiteLocation; l != nil {
		input.SiteLocation = &models.UpdateSiteLocationInput{

			Address:     l.Address,
			CityName:    l.City,
			CountryCode: l.CountryCode,
			StateCode:   l.StateCode,
			Timezone:    l.Timezone,
		}
	}
	_, err := a.Client.SiteUpdateSiteGeneralDetails(ctx, id, input, a.AccountID)
	return apperr.Wrap("Catov2 API SiteUpdateSiteGeneralDetails error", err)
}
func (a SDK) UpdateRange(ctx context.Context, id string, in application.RangeInput) error {
	_, err := a.Client.SiteUpdateNetworkRange(
		ctx,
		id,
		models.UpdateNetworkRangeInput{

			Subnet:           in.Subnet,
			TranslatedSubnet: in.TranslatedSubnet,
			LocalIP:          in.LocalIP,
			MdnsReflector:    in.MdnsReflector,
			Vlan:             in.Vlan,
			DhcpSettings:     dhcp.Input(in.DhcpSettings),
		},
		a.AccountID,
	)
	return apperr.Wrap("Catov2 API SiteUpdateNetworkRange error", err)
}
func (a SDK) UpdateInterface(ctx context.Context, id, index string, in application.InterfaceInput) error {
	input := models.UpdateSocketInterfaceInput{

		DestType: models.SocketInterfaceDestType(in.DestType),
		Name:     in.Name,
	}
	if in.Lan != nil {
		input.Lan = &models.SocketInterfaceLanInput{

			LocalIP:          in.Lan.LocalIP,
			Subnet:           in.Lan.Subnet,
			TranslatedSubnet: in.Lan.TranslatedSubnet,
		}
	}
	if in.Lag != nil {
		input.Lag = &models.SocketInterfaceLagInput{
			MinLinks: in.Lag.MinLinks,
		}
	}
	_, err := a.Client.SiteUpdateSocketInterface(ctx, id, models.SocketInterfaceIDEnum(index), input, a.AccountID)
	return apperr.Wrap("Catov2 API SiteUpdateSocketInterface error", err)
}
func (a SDK) Exchange(ctx context.Context, id, first, second string) error {
	_, err := a.Client.SiteExchangeSocketPorts(
		ctx,
		a.AccountID,
		models.ExchangeSocketPortsInput{

			Site: &models.SiteRefInput{
				By:    models.ObjectRefByID,
				Input: id,
			},
			FirstInterface: &models.SocketInterfaceRefInput{
				InterfaceID: models.SocketInterfaceIDEnum(first),
			},
			SecondInterface: &models.SocketInterfaceRefInput{
				InterfaceID: models.SocketInterfaceIDEnum(second),
			},
		},
	)
	return apperr.Wrap("Catov2 API SiteExchangeSocketPorts error", err)
}
func (a SDK) General(ctx context.Context, id string) (*application.Snapshot, error) {
	result, err := a.Client.SiteGeneralDetails(
		ctx,
		models.SiteRefInput{
			By:    models.ObjectRefByID,
			Input: id,
		},
		a.AccountID,
	)
	if err != nil {
		return nil, apperr.Wrap(fmt.Sprintf("failed to fetch SiteGeneralDetails for site '%s'", id), err)
	}
	if result == nil || result.GetSite().GetSiteGeneralDetails() == nil {
		return nil, nil
	}
	d := result.GetSite().GetSiteGeneralDetails()
	l := d.SiteLocation
	return &application.Snapshot{

		ID:          id,
		Name:        d.GetSite().GetName(),
		SiteType:    (*string)(d.GetSiteType()),
		Description: d.GetDescription(),
		Location: &application.Location{

			CountryCode: l.GetCountryCode(),
			Timezone:    l.GetTimezone(),
			Address:     l.GetAddress(),
			City:        l.GetCityName(),
			StateCode:   l.GetStateCode(),
		},
	}, nil
}
func (a SDK) Configuration(ctx context.Context, id string) (*application.Configuration, error) {
	result, err := a.Client.SiteSocketConfiguration(
		ctx,
		models.SiteSocketConfigurationInput{
			Site: &models.SiteRefInput{
				By:    models.ObjectRefByID,
				Input: id,
			},
		},
		a.AccountID,
	)
	if err != nil {
		return nil, apperr.Wrap(fmt.Sprintf("failed to fetch SiteSocketConfiguration for site '%s'", id), err)
	}
	if result == nil || result.GetSite().GetSiteSocketConfiguration() == nil {
		return nil, apperr.Wrap(fmt.Sprintf("socket configuration for site '%s' was not returned", id), fmt.Errorf(""))
	}
	return ProjectConfiguration(result.GetSite().GetSiteSocketConfiguration()), nil
}
func ProjectConfiguration(in *cato.SiteSocketConfiguration_Site_SiteSocketConfiguration) *application.Configuration {
	if in == nil {
		return nil
	}
	p := in.GetPrimarySocketConfiguration()
	info := p.GetSocketInfo()
	out := &application.Configuration{

		Primary: application.Socket{

			Serial:    p.GetSerial(),
			Platform:  info.GetPlatform(),
			Model:     (*string)(info.GetModel()),
			IsPrimary: info.GetIsPrimary(),
		},
	}
	if s := in.GetSecondarySocketConfiguration(); s != nil {
		info := s.GetSocketInfo()
		out.Secondary = &application.Socket{

			Serial:    s.GetSerial(),
			Platform:  info.GetPlatform(),
			Model:     (*string)(info.GetModel()),
			IsPrimary: info.GetIsPrimary(),
		}
	}
	return out
}
func (a SDK) NativeRange(ctx context.Context, id string) (*application.Range, error) {
	result, err := a.Client.NetworkRangeList(
		ctx,
		a.AccountID,
		models.NetworkRangeListInput{
			Site: &models.SiteRefInput{
				By:    models.ObjectRefByID,
				Input: id,
			},
		},
	)
	if err != nil {
		return nil, apperr.Wrap("Catov2 API NetworkRangeList error", err)
	}
	if result != nil {
		for _, r := range result.GetSite().GetNetworkRangeList().GetItems() {
			if r.RangeType == models.SubnetTypeNative {
				out := &application.Range{

					NetworkRangeID:        r.NetworkRangeID,
					Name:                  r.Name,
					Subnet:                r.Subnet,
					RangeType:             string(r.RangeType),
					LocalIP:               r.LocalIP,
					PrimaryManagementIP:   r.PrimaryManagementIP,
					SecondaryManagementIP: r.SecondaryManagementIP,
					TranslatedSubnet:      r.TranslatedSubnet,
					Gateway:               r.Gateway,
					Vlan:                  r.Vlan,
					MdnsReflector:         r.MdnsReflector,
				}
				if d := r.DhcpSettings; d != nil {
					out.DhcpSettings = dhcp.Snapshot(string(d.DhcpType), d.IPRange, d.RelayGroupID, d.DhcpMicrosegmentation)
				}
				return out, nil
			}
		}
	}
	return nil, apperr.Wrap(
		"Native network range not found",
		fmt.Errorf("no native network range found for site ID %s", id),
	)
}
func (a SDK) DefaultInterface(ctx context.Context, id, stateID string) (*application.Interface, error) {
	result, err := a.Client.EntityLookup(
		ctx,
		a.AccountID,
		models.EntityTypeNetworkInterface,
		new(int64(0)),
		nil,
		&models.EntityInput{
			Type: models.EntityTypeSite,
			ID:   id,
		},
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, apperr.Wrap("Catov2 API EntityLookup 'networkInterface' error", err)
	}
	var found *cato.EntityLookup_EntityLookup_Items
	if result != nil {
		for _, item := range result.EntityLookup.GetItems() {
			if yes, ok := item.HelperFields["isDefault"].(bool); ok && yes {
				found = item
				break
			}
		}
		if found == nil && stateID != "" {
			for _, item := range result.EntityLookup.GetItems() {
				if item.Entity.ID == stateID {
					found = item
					break
				}
			}
		}
	}
	if found == nil {
		return nil, apperr.Wrap(
			"Default interface not found",
			fmt.Errorf("no default network interface found for site ID %s", id),
		)
	}
	field := func(key string) *string {
		if v, ok := found.HelperFields[key].(string); ok {
			return &v
		}
		return nil
	}
	return &application.Interface{

		Index:    field("interfaceId"),
		ID:       &found.Entity.ID,
		Name:     field("interfaceName"),
		DestType: field("destType"),
	}, nil
}
func (a SDK) Exists(ctx context.Context, id string) (bool, error) {
	result, err := a.Client.EntityLookup(
		ctx,
		a.AccountID,
		models.EntityTypeSite,
		nil,
		nil,
		nil,
		nil,
		[]string{
			id,
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return false, apperr.Wrap("Catov2 API error", err)
	}
	return result != nil && len(result.EntityLookup.GetItems()) == 1, nil
}
func (a SDK) Remove(ctx context.Context, id string) error {
	_, err := a.Client.SiteRemoveSite(ctx, id, a.AccountID)
	return apperr.Wrap("Catov2 API SiteRemoveSite error", err)
}
