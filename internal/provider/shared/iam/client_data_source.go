package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// DataConfig holds shared, injectable IAM query configuration.
type DataConfig struct {
	Service *catoclient.Service
	API     RoleReader
}

// Configure receives the provider's configured SDK client.
func (c *DataConfig) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	service, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Invalid provider configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	c.Service = service
}

// Reader returns the injected query boundary or configured SDK.
func (c *DataConfig) Reader() RoleReader {
	if c.API != nil {
		return c.API
	}
	if c.Service == nil || c.Service.Catov2 == nil {
		return nil
	}
	return c.Service.V2()
}

// Account resolves a known data-source account override.
func (c *DataConfig) Account(value types.String) (string, error) {
	if c.Reader() == nil {
		return "", fmt.Errorf("configure the Cato provider before reading IAM data")
	}
	fallback := ""
	if c.Service != nil {
		fallback = c.Service.AccountId
	}
	return Account(value, fallback)
}
