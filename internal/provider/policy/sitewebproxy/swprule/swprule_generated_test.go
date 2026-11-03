package swprule

import (
	"context"
	"testing"

	resource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSiteWebProxyRuleMetadataAndSchema(t *testing.T) {
	t.Parallel()
	r := NewSiteWebProxyRuleResource()
	meta := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, meta)
	if meta.TypeName != "cato_site_web_proxy_rule" {
		t.Fatalf("TypeName = %q, want cato_site_web_proxy_rule", meta.TypeName)
	}
	schemaResp := &resource.SchemaResponse{}
	r.(interface {
		Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	}).Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	for _, attr := range []string{"at", "rule"} {
		if schemaResp.Schema.Attributes[attr] == nil {
			t.Fatalf("schema missing %s attribute", attr)
		}
	}
}
