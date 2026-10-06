package application

import (
	"context"

	apperr "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
)

type Settings struct {
	DhcpType              string
	IPRange               *string
	RelayGroupID          *string
	RelayGroupName        *string
	DhcpMicrosegmentation *bool
}

type RelayLookup interface {
	RelayID(context.Context, string) (string, error)
	RelayName(context.Context, string) (*string, error)
}

// Prepare preserves the existing payload rules, including omission of a configured relay ID.
func Prepare(ctx context.Context, lookup RelayLookup, rangeType string, settings *Settings) (*Settings, error) {
	if settings == nil || (rangeType != "Native" && rangeType != "VLAN") || settings.DhcpType == "" {
		return nil, nil
	}
	input := &Settings{
		DhcpType: settings.DhcpType,
		IPRange:  settings.IPRange,
	}
	if input.DhcpType == "DHCP_RELAY" && settings.RelayGroupID == nil {
		if settings.RelayGroupName == nil {
			return nil, apperr.New(
				"Missing DHCP relay group name",
				"DHCP settings of type DHCP_RELAY require a relay group name to be specified.",
			)
		}
		id, err := lookup.RelayID(ctx, *settings.RelayGroupName)
		if err != nil {
			return nil, err
		}
		input.RelayGroupID = &id
	}
	if input.DhcpType == "DHCP_RANGE" {
		input.DhcpMicrosegmentation = settings.DhcpMicrosegmentation
	}
	return input, nil
}

func ResolveName(ctx context.Context, lookup RelayLookup, settings *Settings) error {
	if settings != nil && settings.DhcpType == "DHCP_RELAY" && settings.RelayGroupID != nil {
		name, err := lookup.RelayName(ctx, *settings.RelayGroupID)
		if err != nil {
			return err
		}
		settings.RelayGroupName = name
	}
	return nil
}
