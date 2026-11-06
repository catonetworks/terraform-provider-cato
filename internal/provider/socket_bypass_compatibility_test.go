package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

func TestSocketBypassPreservesExistingSchema(t *testing.T) {
	ctx := context.Background()
	for _, factory := range (&catoProvider{}).Resources(ctx) {
		r := factory()
		var metadata resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
		if metadata.TypeName != "cato_socket_bypass_rule" {
			continue
		}
		var response resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &response)
		rule := response.Schema.Attributes["rule"].(schema.SingleNestedAttribute)
		action := rule.Attributes["action"].(schema.SingleNestedAttribute)
		event, ok := action.Attributes["event"].(schema.SingleNestedAttribute)
		require.True(t, ok, "preserve action.event instead of silently ignoring existing event configuration")
		require.Contains(t, event.Attributes, "enabled")
		service := rule.Attributes["service"].(schema.SingleNestedAttribute)
		_, ok = service.Attributes["simple"].(schema.SetNestedAttribute)
		require.True(t, ok, "preserve the set of named service objects")
		_, ok = rule.Attributes["exception"].(schema.SetNestedAttribute)
		require.True(t, ok, "preserve exception set state")
		require.Contains(t, rule.Attributes, "index")
		require.Contains(t, rule.Attributes, "section")
		require.False(t, response.Schema.Attributes["at"].IsRequired())
		return
	}
	t.Fatal("Socket Bypass rule is not registered")
}
