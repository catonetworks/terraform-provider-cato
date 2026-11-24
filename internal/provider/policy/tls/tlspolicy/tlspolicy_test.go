package tlspolicy

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTLSInspectPolicyMetadataAndSchema(t *testing.T) {
	t.Parallel()
	r := NewTLSInspectPolicyResource()
	meta := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, meta)
	if meta.TypeName != "cato_tls_inspect_policy" {
		t.Fatalf("TypeName = %q, want cato_tls_inspect_policy", meta.TypeName)
	}
	schemaResp := &resource.SchemaResponse{}
	r.(interface {
		Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	}).Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	if schemaResp.Schema.Attributes["enabled"] == nil {
		t.Fatal("schema missing enabled attribute")
	}
}

func TestTLSInspectPolicyUpdateInputMapsToggle(t *testing.T) {
	t.Parallel()
	in := convertUpdatePolicyValues(TLSInspectPolicyModel{Enabled: types.BoolValue(true)})
	if in.State == nil || *in.State != cato_models.PolicyToggleStateEnabled {
		t.Fatalf("enabled=true mapped to %+v, want ENABLED", in.State)
	}
	in = convertUpdatePolicyValues(TLSInspectPolicyModel{Enabled: types.BoolValue(false)})
	if in.State == nil || *in.State != cato_models.PolicyToggleStateDisabled {
		t.Fatalf("enabled=false mapped to %+v, want DISABLED", in.State)
	}
}

func TestTLSInspectPolicyReadStateID(t *testing.T) {
	t.Parallel()
	if got := convertReadStateValues().ID.ValueString(); got != "tls_inspect" {
		t.Fatalf("ID = %q, want tls_inspect", got)
	}
}
