package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/refactor/internal/provider/resources"
)

func TestProviderConfigureWrapsSDK(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := NewProvider("test")()
	var schemaResponse provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResponse)
	value, diagnostics := types.ObjectValueFrom(ctx, schemaResponse.Schema.Type().(types.ObjectType).AttrTypes, configModel{
		BaseURL: types.StringValue("https://cato.invalid/graphql"), Token: types.StringValue("test-only"), AccountID: types.StringValue("account-123"),
	})
	require.False(t, diagnostics.HasError(), diagnostics)
	raw, err := value.ToTerraformValue(ctx)
	require.NoError(t, err)
	var response provider.ConfigureResponse
	p.Configure(ctx, provider.ConfigureRequest{Config: tfsdk.Config{Schema: schemaResponse.Schema, Raw: raw}}, &response)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	deps, ok := response.ResourceData.(*resources.Dependencies)
	require.True(t, ok)
	require.NotNil(t, deps.CatoAPI)
	require.NotNil(t, deps.ResourceLock)
	require.Equal(t, "account-123", deps.AccountID)
}
