package application

import (
	"context"

	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
)

type Service struct {
	Port  SocketSitePort
	Retry Retrier
	// ValidateSnapshot allows the caller to report projection errors inside the hydration retry boundary.
	ValidateSnapshot func(*Snapshot) error
}

func (s Service) Create(ctx context.Context, in Input) (Result, error) {
	id, err := s.Port.Add(ctx, in.Add)
	if err != nil {
		return Result{}, err
	}
	rangeData, err := s.Port.NativeRange(ctx, id)
	if err != nil {
		return Result{}, err
	}
	if err = s.updateRange(ctx, rangeData.NetworkRangeID, in.Range); err != nil {
		return Result{}, err
	}
	if err = s.exchange(ctx, id, DefaultInterfaceIndex(in.ConnectionType), in.DesiredIndex); err != nil {
		return Result{}, err
	}
	if err = s.updateInterface(ctx, id, in); err != nil {
		return Result{}, err
	}
	in.Read.ID = id
	result, err := s.Retry.Run(ctx, CreateHydrationPolicy, func() (Result, error) { return s.Read(ctx, in.Read) })
	if err != nil {
		return Result{}, err
	}
	result.ID = id
	result.Pending = !result.Found
	return result, nil
}
func (s Service) Update(ctx context.Context, in Input) (Result, error) {
	if err := s.Port.UpdateGeneral(ctx, in.ID, in.General); err != nil {
		return Result{}, err
	}
	if err := s.updateRange(ctx, in.RangeID, in.Range); err != nil {
		return Result{}, err
	}
	if err := s.exchange(ctx, in.ID, in.CurrentIndex, in.DesiredIndex); err != nil {
		return Result{}, err
	}
	if err := s.updateInterface(ctx, in.ID, in); err != nil {
		return Result{}, err
	}
	return s.Read(ctx, in.Read)
}
func (s Service) updateRange(ctx context.Context, id string, in RangeInput) error {
	settings, err := dhcp.Prepare(ctx, s.Port, "Native", in.DhcpSettings)
	if err != nil {
		return err
	}
	in.DhcpSettings = settings
	return s.Port.UpdateRange(ctx, id, in)
}
func (s Service) exchange(ctx context.Context, id, current string, desired *string) error {
	if desired == nil || current == *desired {
		return nil
	}
	return s.Port.Exchange(ctx, id, current, *desired)
}
func (s Service) updateInterface(ctx context.Context, id string, in Input) error {
	if in.IsHA && in.ConnectionType == "SOCKET_GCP1500" {
		return nil
	}
	return s.Port.UpdateInterface(ctx, id, in.InterfaceIndex, in.Interface)
}
func (s Service) Read(ctx context.Context, in ReadInput) (Result, error) {
	snapshot, err := s.Port.General(ctx, in.ID)
	if err != nil {
		return Result{}, err
	}
	if snapshot == nil {
		return Result{
			ID: in.ID,
		}, nil
	}
	snapshot.Configuration, err = s.Port.Configuration(ctx, in.ID)
	if err != nil {
		return Result{}, err
	}
	snapshot.Range, err = s.Port.NativeRange(ctx, in.ID)
	if err != nil {
		return Result{}, err
	}
	snapshot.Interface, err = s.Port.DefaultInterface(ctx, in.ID, in.InterfaceID)
	if err != nil {
		return Result{}, err
	}
	if in.ResolveRelayName {
		if err = dhcp.ResolveName(ctx, s.Port, snapshot.Range.DhcpSettings); err != nil {
			return Result{}, err
		}
	}
	if s.ValidateSnapshot != nil {
		if err = s.ValidateSnapshot(snapshot); err != nil {
			return Result{}, err
		}
	}
	return Result{
		ID:       in.ID,
		Snapshot: snapshot,
		Found:    true,
	}, nil
}
func (s Service) Delete(ctx context.Context, id string) error {
	exists, err := s.Port.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return s.Port.Remove(ctx, id)
}
func DefaultInterfaceIndex(connection string) string {
	switch connection {
	case "SOCKET_AWS1500", "SOCKET_AZ1500", "SOCKET_ESX1500", "SOCKET_GCP1500", "SOCKET_X1500":
		return "LAN1"
	case "SOCKET_X1600", "SOCKET_X1600_LTE":
		return "INT_5"
	case "SOCKET_X1700":
		return "INT_3"
	default:
		return ""
	}
}
