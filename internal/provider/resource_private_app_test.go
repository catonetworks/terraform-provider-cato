package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPrivateAppProbingSchemaAcceptsAPIDefault(t *testing.T) {
	t.Parallel()

	r := &privateAppResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["private_app_probing"]
	if !ok {
		t.Fatal("expected private_app_probing attribute")
	}

	probing, ok := attr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected schema.SingleNestedAttribute, got %T", attr)
	}

	if !probing.Optional {
		t.Fatal("expected private_app_probing to be optional")
	}
	if !probing.Computed {
		t.Fatal("expected private_app_probing to accept API defaults")
	}
	if len(probing.PlanModifiers) == 0 {
		t.Fatal("expected private_app_probing to preserve API defaults during planning")
	}
}

func TestPrivateAppComputedAttributesPreserveStateWhenUnknown(t *testing.T) {
	t.Parallel()

	r := &privateAppResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	for _, attrName := range []string{"creation_time", "id"} {
		attr, ok := resp.Schema.Attributes[attrName].(schema.StringAttribute)
		if !ok {
			t.Fatalf("expected %s to be schema.StringAttribute", attrName)
		}
		if len(attr.PlanModifiers) == 0 {
			t.Fatalf("expected %s to preserve state during planning", attrName)
		}
	}

	attr, ok := resp.Schema.Attributes["published_app_domain"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected published_app_domain to be schema.SingleNestedAttribute")
	}
	for _, attrName := range []string{"creation_time", "id"} {
		computedAttr, ok := attr.Attributes[attrName].(schema.StringAttribute)
		if !ok {
			t.Fatalf("expected published_app_domain.%s to be schema.StringAttribute", attrName)
		}
		if len(computedAttr.PlanModifiers) == 0 {
			t.Fatalf("expected published_app_domain.%s to preserve state during planning", attrName)
		}
	}
}

func TestPreparePrivateAppProbingSkipsUnconfiguredValue(t *testing.T) {
	t.Parallel()

	r := &privateAppResource{}
	diags := diag.Diagnostics{}

	got := r.preparePrivateAppProbing(
		context.Background(),
		types.ObjectUnknown(PrivateAppProbingTypes),
		&diags,
	)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got != nil {
		t.Fatalf("expected omitted probing input, got %#v", got)
	}
}
