package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/spf13/cast"

	apperr "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/client"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/adapter"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/netrange/application"
)

type NetworkRangeClient interface {
	dhcp.Client
	SiteAddNetworkRange(
		context.Context,
		string,
		models.AddNetworkRangeInput,
		string,
		...clientv2.RequestInterceptor,
	) (
		*cato.SiteAddNetworkRange,
		error,
	)
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
	SiteRemoveNetworkRange(context.Context, string, string, ...clientv2.RequestInterceptor) (*cato.SiteRemoveNetworkRange, error)
	NetworkRange(context.Context, string, string, ...clientv2.RequestInterceptor) (*cato.NetworkRange, error)
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
}
type SDK struct {
	Client    NetworkRangeClient
	AccountID string
}

var _ application.NetworkRangePort = SDK{}

func ResolveClient(injected NetworkRangeClient, data *client.CatoClientData) NetworkRangeClient {
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
func (a SDK) Add(ctx context.Context, iface string, in application.Input) (string, error) {
	result, err := a.Client.SiteAddNetworkRange(ctx, iface, models.AddNetworkRangeInput{

		Name:             value(in.Name),
		RangeType:        models.SubnetType(value(in.RangeType)),
		Subnet:           value(in.Subnet),
		Gateway:          in.Gateway,
		LocalIP:          in.LocalIP,
		TranslatedSubnet: in.TranslatedSubnet,
		InternetOnly:     in.InternetOnly,
		MdnsReflector:    in.MdnsReflector,
		Vlan:             in.Vlan,
		DhcpSettings: dhcp.Input(
			in.DhcpSettings,
		),
	}, a.AccountID)
	if err != nil {
		return "", apperr.Wrap("Cato API SiteAddNetworkRange Error", err)
	}
	if result == nil || result.Site.AddNetworkRange == nil {
		return "", apperr.Wrap(
			"Cato API SiteAddNetworkRange Error",
			fmt.Errorf("empty network range response returned from API"),
		)
	}
	return result.Site.AddNetworkRange.NetworkRangeID, nil
}
func (a SDK) Update(ctx context.Context, in application.Input) error {
	_, err := a.Client.SiteUpdateNetworkRange(ctx, in.ID, models.UpdateNetworkRangeInput{

		Name: in.Name,
		RangeType: enum(
			in.RangeType,
		),
		Subnet:           in.Subnet,
		Gateway:          in.Gateway,
		LocalIP:          in.LocalIP,
		TranslatedSubnet: in.TranslatedSubnet,
		InternetOnly:     in.InternetOnly,
		MdnsReflector:    in.MdnsReflector,
		Vlan:             in.Vlan,
		DhcpSettings: dhcp.Input(
			in.DhcpSettings,
		),
	}, a.AccountID)
	return mutationError(err)
}
func (a SDK) Remove(ctx context.Context, id string) error {
	_, err := a.Client.SiteRemoveNetworkRange(ctx, id, a.AccountID)
	return mutationError(err)
}
func mutationError(err error) error {
	if err == nil {
		return nil
	}
	var response cato.RespErrors
	if json.Unmarshal([]byte(err.Error()), &response) == nil && len(response.GraphQLErrors) > 0 {
		msg := response.GraphQLErrors[0].Message
		if strings.Contains(msg, "Network range with id: ") && strings.Contains(msg, "is not found") {
			return apperr.ErrAbsent
		}
	}
	return apperr.Wrap("Catov2 API error", err)
}
func (a SDK) Fetch(ctx context.Context, id string) (*application.Snapshot, error) {
	result, err := a.Client.NetworkRange(ctx, a.AccountID, id)
	if err != nil {
		if gql, ok := errors.AsType[*clientv2.ErrorResponse](err); ok && gql.GqlErrors != nil && len(*gql.GqlErrors) > 0 &&
			strings.Contains((*gql.GqlErrors)[0].Message, "Invalid network range id: ") {
			return nil, nil
		}
		return nil, apperr.Wrap(
			"Catov2 NetworkRange API error",
			fmt.Errorf("error fetching network range details for range ID '%s': %w", id, err),
		)
	}
	if result == nil || result.Site.NetworkRange == nil {
		return nil, nil
	}
	r := result.Site.NetworkRange
	out := &application.Snapshot{

		NetworkRangeID:   id,
		Name:             r.Name,
		RangeType:        string(r.RangeType),
		Subnet:           r.Subnet,
		Gateway:          r.Gateway,
		LocalIP:          r.LocalIP,
		TranslatedSubnet: r.TranslatedSubnet,
		InternetOnly:     r.InternetOnly,
		MdnsReflector:    r.MdnsReflector,
		Vlan:             r.Vlan,
	}
	if d := r.DhcpSettings; d != nil {
		out.DhcpSettings = dhcp.Snapshot(string(d.DhcpType), d.IPRange, d.RelayGroupID, d.DhcpMicrosegmentation)
	}
	return out, nil
}

var numberRE = regexp.MustCompile(`^\d+$`)

func normalize(index string) string {
	if numberRE.MatchString(index) {
		return "INT_" + index
	}
	return index
}
func (a SDK) interfaces(ctx context.Context, site string, ids []string) (*cato.EntityLookup, error) {
	return a.Client.EntityLookup(
		ctx,
		a.AccountID,
		models.EntityTypeNetworkInterface,
		new(int64(0)),
		nil,
		&models.EntityInput{
			Type: models.EntityTypeSite,
			ID:   site,
		},
		nil,
		ids,
		nil,
		nil,
		nil,
	)
}
func (a SDK) InterfaceID(ctx context.Context, site, index string) (string, error) {
	result, err := a.interfaces(ctx, site, nil)
	if err != nil {
		return "", apperr.Wrap("Error retrieving network interface "+index, err)
	}
	if result != nil {
		for _, item := range result.GetEntityLookup().GetItems() {
			if normalize(cast.ToString(item.GetHelperFields()["interfaceId"])) == index {
				return item.GetEntity().GetID(), nil
			}
		}
	}
	return "", apperr.Wrap(
		"Error retrieving network interface",
		fmt.Errorf("network interface with index '%s' not found in site '%s'", index, site),
	)
}
func (a SDK) InterfaceIndex(ctx context.Context, site, id string) (string, error) {
	result, err := a.interfaces(ctx, site, []string{
		id,
	})
	if err != nil {
		return "", apperr.Wrap("Error retrieving network interface "+id, err)
	}
	if result != nil {
		for _, item := range result.GetEntityLookup().GetItems() {
			if item.GetEntity().GetID() == id {
				return normalize(cast.ToString(item.GetHelperFields()["interfaceId"])), nil
			}
		}
	}
	return "", apperr.Wrap(
		"Error retrieving network interface",
		fmt.Errorf("network interface with id '%s' not found in site '%s'", id, site),
	)
}
func (a SDK) SiteID(ctx context.Context, id string) (siteID, interfaceName string, err error) {
	result, err := a.Client.EntityLookup(
		ctx,
		a.AccountID,
		models.EntityTypeSiteRange,
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
		return "", "", fmt.Errorf("failed to lookup network range site: %w", err)
	}
	if result == nil || len(result.EntityLookup.GetItems()) == 0 {
		return "", "", fmt.Errorf("network range %s not found in entityLookup", id)
	}
	fields := result.EntityLookup.GetItems()[0].GetHelperFields()
	if fields == nil {
		return "", "", fmt.Errorf("no helperFields returned for network range %s", id)
	}
	site := cast.ToString(fields["siteId"])
	if site == "" {
		return "", "", fmt.Errorf("siteId not found in helperFields for network range %s", id)
	}
	return site, cast.ToString(fields["interfaceName"]), nil
}
func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func enum(v *string) *models.SubnetType {
	if v == nil {
		return nil
	}
	r := models.SubnetType(*v)
	return &r
}
