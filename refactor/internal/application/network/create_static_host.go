package network

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/application/network/entities"
)

type CreateStaticHostAdapter interface {
	CreateStaticHost(ctx context.Context, accountID, siteID string, host entities.StaticHost) (string, error)
}

type CreateStaticHostUseCase struct {
	adapter CreateStaticHostAdapter
}

type CreateStaticHostCommand struct {
	AccountID string
	SiteID    string
	Host      entities.StaticHost
}

func NewCreateStaticHostUseCase(adapter CreateStaticHostAdapter) (*CreateStaticHostUseCase, error) {
	if adapter == nil {
		return nil, errors.New("static-host adapter is required")
	}
	return &CreateStaticHostUseCase{adapter: adapter}, nil
}

func (u *CreateStaticHostUseCase) Execute(ctx context.Context, command CreateStaticHostCommand) (string, error) {
	for _, field := range []struct{ name, value string }{
		{"account_id", command.AccountID}, {"site_id", command.SiteID},
		{"name", command.Host.Name}, {"ip", command.Host.IP},
	} {
		if strings.TrimSpace(field.value) == "" {
			return "", &ValidationError{Field: field.name}
		}
	}
	if _, err := netip.ParseAddr(command.Host.IP); err != nil {
		return "", &ValidationError{Field: "ip"}
	}
	if command.Host.MacAddress != nil {
		if _, err := net.ParseMAC(*command.Host.MacAddress); err != nil {
			return "", &ValidationError{Field: "mac_address"}
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	// Creation is a single mutation. Do not replay an uncertain result.
	id, err := u.adapter.CreateStaticHost(ctx, command.AccountID, command.SiteID, command.Host)
	if err != nil {
		return "", classifyCatoAPIError(err)
	}
	return id, nil
}
