package application

import (
	"context"

	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
)

type NetworkRangePort interface {
	dhcp.RelayLookup
	Add(context.Context, string, Input) (string, error)
	Update(context.Context, Input) error
	Remove(context.Context, string) error
	Fetch(context.Context, string) (*Snapshot, error)
	InterfaceID(context.Context, string, string) (string, error)
	InterfaceIndex(context.Context, string, string) (string, error)
}
