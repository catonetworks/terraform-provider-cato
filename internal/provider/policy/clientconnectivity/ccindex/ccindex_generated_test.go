package ccindex

import (
	"context"
	"testing"

	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCcRulesIndexMetadataAndSchema(t *testing.T) {
	t.Parallel()
	r := NewCcRulesIndexResource()
	meta := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, meta)
	if meta.TypeName != "cato_bulk_cc_move_rule" {
		t.Fatalf("TypeName = %q, want cato_bulk_cc_move_rule", meta.TypeName)
	}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	if schemaResp.Schema.Attributes == nil || schemaResp.Schema.Attributes["section_data"] == nil || schemaResp.Schema.Attributes["rule_data"] == nil {
		t.Fatal("schema missing section_data/rule_data")
	}
}

func TestCcDecodeOrderingInputs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sectionEntryType := types.ObjectType{AttrTypes: map[string]attr.Type{"section_index": types.Int64Type}}
	ruleEntryType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"index_in_section": types.Int64Type,
		"section_name":     types.StringType,
	}}
	sectionData, d := types.MapValue(sectionEntryType, map[string]attr.Value{
		"Section A": types.ObjectValueMust(sectionEntryType.AttrTypes, map[string]attr.Value{
			"section_index": types.Int64Value(1),
		}),
	})
	if d.HasError() {
		t.Fatalf("section map: %v", d)
	}
	ruleData, d := types.MapValue(ruleEntryType, map[string]attr.Value{
		"Rule 1": types.ObjectValueMust(ruleEntryType.AttrTypes, map[string]attr.Value{
			"index_in_section": types.Int64Value(2),
			"section_name":     types.StringValue("Section A"),
		}),
	})
	if d.HasError() {
		t.Fatalf("rule map: %v", d)
	}

	plan := CcRulesIndex{SectionData: sectionData, RuleData: ruleData}
	sections, rules, err := decodeCcOrderingInputs(ctx, plan)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if sections["Section A"] != 1 {
		t.Fatalf("sections = %+v", sections)
	}
	row := rules["Rule 1"]
	if row.IndexInSection != 2 || row.SectionName != "Section A" {
		t.Fatalf("rules = %+v", rules)
	}
}
