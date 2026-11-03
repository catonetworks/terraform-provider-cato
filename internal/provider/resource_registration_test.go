package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Validate registration and schema compatibility together: policy families can
// compile independently while still colliding when registered in the provider.
func TestRegisteredResourceSchemas(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := &catoProvider{}
	seen := map[string]bool{}
	for _, factory := range p.Resources(ctx) {
		r := factory()
		var metadata resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
		if seen[metadata.TypeName] {
			t.Fatalf("resource %s registered twice", metadata.TypeName)
		}
		seen[metadata.TypeName] = true
		t.Run(metadata.TypeName, func(t *testing.T) {
			var response resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &response)
			if response.Diagnostics.HasError() {
				t.Fatalf("schema diagnostics: %v", response.Diagnostics)
			}
			if diagnostics := response.Schema.ValidateImplementation(ctx); diagnostics.HasError() {
				t.Fatalf("invalid schema: %v", diagnostics)
			}
		})
	}
	for _, name := range []string{"cato_always_on_rule_bulk", "cato_always_on_policy"} {
		if !seen[name] {
			t.Errorf("canonical Always On resource %s is missing", name)
		}
	}
	for _, name := range []string{"cato_bulk_ao_move_rule", "cato_ztna_always_on_policy"} {
		if seen[name] {
			t.Errorf("duplicate Always On resource %s must not be registered", name)
		}
	}
}
