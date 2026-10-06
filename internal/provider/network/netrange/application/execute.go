package application

import (
	"context"
	"errors"

	apperr "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
)

type Service struct{ Port NetworkRangePort }

func (s Service) Create(ctx context.Context, input Input, read ReadInput) (Result, error) {
	settings, err := dhcp.Prepare(ctx, s.Port, value(input.RangeType), input.DhcpSettings)
	if err != nil {
		return Result{}, err
	}
	input.DhcpSettings = settings
	id := value(input.InterfaceID)
	if input.InterfaceID == nil {
		id, err = s.Port.InterfaceID(ctx, input.SiteID, input.InterfaceIndex)
		if err != nil {
			return Result{}, err
		}
	}
	rangeID, err := s.Port.Add(ctx, id, input)
	if err != nil {
		return Result{}, err
	}
	read.ID = rangeID
	return s.Read(ctx, read)
}
func (s Service) Update(ctx context.Context, input Input, read ReadInput) (Result, error) {
	settings, err := dhcp.Prepare(ctx, s.Port, value(input.RangeType), input.DhcpSettings)
	if err != nil {
		return Result{}, err
	}
	input.DhcpSettings = settings
	if err = s.Port.Update(ctx, input); err != nil {
		if errors.Is(err, apperr.ErrAbsent) {
			return Result{}, nil
		}
		return Result{}, err
	}
	return s.Read(ctx, read)
}
func (s Service) Read(ctx context.Context, in ReadInput) (Result, error) {
	snapshot, err := s.Port.Fetch(ctx, in.ID)
	if err != nil {
		return Result{}, err
	}
	if snapshot == nil {
		return Result{}, nil
	}
	snapshot.InterfaceID, snapshot.InterfaceIndex = in.InterfaceID, in.InterfaceIndex
	if in.ResolveInterface {
		if in.InterfaceByID {
			snapshot.InterfaceIndex, err = s.Port.InterfaceIndex(ctx, in.SiteID, in.InterfaceID)
		} else {
			snapshot.InterfaceID, err = s.Port.InterfaceID(ctx, in.SiteID, in.InterfaceIndex)
		}
		if err != nil {
			return Result{}, err
		}
	}
	if in.ResolveRelayName {
		if err = dhcp.ResolveName(ctx, s.Port, snapshot.DhcpSettings); err != nil {
			return Result{}, err
		}
	}
	return Result{
		Snapshot: snapshot,
		Found:    true,
	}, nil
}
func (s Service) Delete(ctx context.Context, id string) error {
	err := s.Port.Remove(ctx, id)
	if errors.Is(err, apperr.ErrAbsent) {
		return nil
	}
	return err
}
func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
