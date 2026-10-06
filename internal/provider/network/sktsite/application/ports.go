package application

import (
	"context"

	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
)

type SocketSitePort interface {
	dhcp.RelayLookup
	Add(context.Context, AddInput) (string, error)
	UpdateGeneral(context.Context, string, GeneralInput) error
	UpdateRange(context.Context, string, RangeInput) error
	Exchange(context.Context, string, string, string) error
	UpdateInterface(context.Context, string, string, InterfaceInput) error
	General(context.Context, string) (*Snapshot, error)
	Configuration(context.Context, string) (*Configuration, error)
	NativeRange(context.Context, string) (*Range, error)
	DefaultInterface(context.Context, string, string) (*Interface, error)
	Exists(context.Context, string) (bool, error)
	Remove(context.Context, string) error
}
type Retrier interface {
	Run(context.Context, RetryPolicy, func() (Result, error)) (Result, error)
}
