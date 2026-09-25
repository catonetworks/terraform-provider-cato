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
