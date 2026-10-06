package application

import (
	"time"

	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
)

type Location struct {
	Address, City, StateCode *string
	CountryCode, Timezone    string
}
type AddInput struct {
	ConnectionType, Name, NativeNetworkRange, SiteType string
	Description, TranslatedSubnet                      *string
	SiteLocation                                       *Location
	Vlan                                               *int64
}
type GeneralLocation struct {
	Address, City, StateCode, CountryCode, Timezone *string
}

type GeneralInput struct {
	Name, Description, SiteType *string
	SiteLocation                *GeneralLocation
}
type RangeInput struct {
	Subnet, TranslatedSubnet, LocalIP *string
	MdnsReflector                     *bool
	Vlan                              *int64
	DhcpSettings                      *dhcp.Settings
}
type LanInput struct {
	LocalIP, Subnet  string
	TranslatedSubnet *string
}
type LagInput struct{ MinLinks int64 }
type InterfaceInput struct {
	DestType string
	Name     *string
	Lan      *LanInput
	Lag      *LagInput
}
type Input struct {
	ID, RangeID, CurrentIndex, InterfaceIndex, ConnectionType string
	DesiredIndex                                              *string
	IsHA                                                      bool
	Add                                                       AddInput
	General                                                   GeneralInput
	Range                                                     RangeInput
	Interface                                                 InterfaceInput
	Read                                                      ReadInput
}
type ReadInput struct {
	ID, InterfaceID  string
	ResolveRelayName bool
}
type Range struct {
	NetworkRangeID, Subnet, Name, RangeType                                        string
	LocalIP, PrimaryManagementIP, SecondaryManagementIP, TranslatedSubnet, Gateway *string
	Vlan                                                                           *int64
	MdnsReflector                                                                  bool
	DhcpSettings                                                                   *dhcp.Settings
}
type Interface struct{ Index, ID, Name, DestType *string }
type Socket struct {
	Serial, Platform, Model *string
	IsPrimary               bool
}
type Configuration struct {
	Primary   Socket
	Secondary *Socket
}
type Snapshot struct {
	ID, Name              string
	SiteType, Description *string
	Location              *Location
	Configuration         *Configuration
	Range                 *Range
	Interface             *Interface
}
type Result struct {
	ID             string
	Snapshot       *Snapshot
	Found, Pending bool
}
type RetryPolicy struct {
	Attempts int
	Delay    time.Duration
}

var CreateHydrationPolicy = RetryPolicy{
	Attempts: 6,
	Delay:    2 * time.Second,
}
