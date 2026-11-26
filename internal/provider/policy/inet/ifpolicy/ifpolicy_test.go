package ifpolicy

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

func TestInternetFirewallPolicyMetadataAndSchema(t *testing.T) {
	t.Parallel()
	r := NewInternetFirewallPolicyResource()
	meta := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, meta)
	if meta.TypeName != "cato_if_policy" {
		t.Fatalf("TypeName = %q, want cato_if_policy", meta.TypeName)
	}
	schemaResp := &resource.SchemaResponse{}
	r.(interface {
		Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	}).Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	if schemaResp.Schema.Attributes["enabled"] == nil {
		t.Fatal("schema missing enabled attribute")
	}
}

func TestInternetFirewallPolicyUpdateInputMapsToggle(t *testing.T) {
	t.Parallel()
	in := convertUpdatePolicyValues(InternetFirewallPolicyModel{Enabled: types.BoolValue(true)})
	if in.State == nil || *in.State != cato_models.PolicyToggleStateEnabled {
		t.Fatalf("enabled=true mapped to %+v, want ENABLED", in.State)
	}
	in = convertUpdatePolicyValues(InternetFirewallPolicyModel{Enabled: types.BoolValue(false)})
	if in.State == nil || *in.State != cato_models.PolicyToggleStateDisabled {
		t.Fatalf("enabled=false mapped to %+v, want DISABLED", in.State)
	}
}

func TestInternetFirewallPolicyReadStateID(t *testing.T) {
	t.Parallel()
	if got := convertReadStateValues().ID.ValueString(); got != "internet_firewall" {
		t.Fatalf("ID = %q, want internet_firewall", got)
	}
}
