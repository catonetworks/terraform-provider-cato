package application

import dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"

type Input struct {
	ID, SiteID, InterfaceIndex         string
	InterfaceID                        *string
	Name, RangeType, Subnet            *string
	Gateway, LocalIP, TranslatedSubnet *string
	InternetOnly, MdnsReflector        *bool
	Vlan                               *int64
	DhcpSettings                       *dhcp.Settings
}
type ReadInput struct {
	ID, SiteID, InterfaceID, InterfaceIndex           string
	ResolveInterface, InterfaceByID, ResolveRelayName bool
}
type Snapshot struct {
	NetworkRangeID, Name, RangeType, Subnet string
	Gateway, LocalIP, TranslatedSubnet      *string
	InternetOnly, MdnsReflector             bool
	Vlan                                    *int64
	DhcpSettings                            *dhcp.Settings
	InterfaceID, InterfaceIndex             string
}
type Result struct {
	Snapshot *Snapshot
	Found    bool
}
