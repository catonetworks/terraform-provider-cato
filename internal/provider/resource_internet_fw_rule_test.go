package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// exceptionSourceUserNameValidators returns validators on rule.exceptions[].source.user[].name
// from the live cato_if_rule schema. Fails the test if the schema shape changes.
func exceptionSourceUserNameValidators(t *testing.T, s schema.Schema) []validator.String {
	t.Helper()

	rule, ok := s.Attributes["rule"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal(`schema: expected top-level "rule" to be SingleNestedAttribute`)
	}
	exceptions, ok := rule.Attributes["exceptions"].(schema.SetNestedAttribute)
	if !ok {
		t.Fatal(`schema: expected rule.exceptions to be SetNestedAttribute`)
	}
	src, ok := exceptions.NestedObject.Attributes["source"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal(`schema: expected exception source to be SingleNestedAttribute`)
	}
	user, ok := src.Attributes["user"].(schema.SetNestedAttribute)
	if !ok {
		t.Fatal(`schema: expected exception source.user to be SetNestedAttribute`)
	}
	nameAttr, ok := user.NestedObject.Attributes["name"].(schema.StringAttribute)
	if !ok {
		t.Fatal(`schema: expected exception source.user name to be StringAttribute`)
	}
	return nameAttr.Validators
}

// minimalUserSetSchema is a root schema with only attribute "user" matching the nested object
// shape of exception source.user (name + id). Used with tfsdk.Config so ConflictsWith path
// resolution matches production (sibling id under the same set element).
func minimalUserSetSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user": schema.SetNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{Optional: true, Computed: true},
						"id":   schema.StringAttribute{Optional: true, Computed: true},
					},
				},
			},
		},
	}
}

// mustUserSetConfig builds one exception-style source.user set element and a tfsdk.Config whose
// Raw value contains that set under root attribute "user".
func mustUserSetConfig(t *testing.T, ctx context.Context, nameVal, idVal attr.Value) (userElem types.Object, cfg tfsdk.Config) {
	t.Helper()
	userObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"id":   types.StringType,
		},
	}
	userElem, diags := types.ObjectValue(userObjType.AttrTypes, map[string]attr.Value{
		"name": nameVal,
		"id":   idVal,
	})
	if diags.HasError() {
		t.Fatalf("user element object: %s", diags)
	}
	setVal, diags := types.SetValue(userObjType, []attr.Value{userElem})
	if diags.HasError() {
		t.Fatalf("user set: %s", diags)
	}
	root, diags := types.ObjectValue(
		map[string]attr.Type{"user": types.SetType{ElemType: userObjType}},
		map[string]attr.Value{"user": setVal},
	)
	if diags.HasError() {
		t.Fatalf("root object: %s", diags)
	}
	raw, err := root.ToTerraformValue(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return userElem, tfsdk.Config{Raw: raw, Schema: minimalUserSetSchema()}
}

func validateExceptionUserName(
	t *testing.T,
	ctx context.Context,
	cfg tfsdk.Config,
	userElem types.Object,
	nameStr types.String,
	validators []validator.String,
) validator.StringResponse {
	t.Helper()
	attrPath := path.Root("user").AtSetValue(userElem).AtName("name")
	pathExpr := attrPath.Expression()

	var combined validator.StringResponse
	for _, v := range validators {
		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			Path:           attrPath,
			PathExpression: pathExpr,
			Config:         cfg,
			ConfigValue:    nameStr,
		}, &resp)
		combined.Diagnostics.Append(resp.Diagnostics...)
	}
	return combined
}

func TestInternetFwRuleSchema_ExceptionSourceUser_hasNameIdValidators(t *testing.T) {
	ctx := context.Background()
	var r internetFwRuleResource
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	validators := exceptionSourceUserNameValidators(t, schemaResp.Schema)
	if len(validators) == 0 {
		t.Fatal(`expected rule.exceptions.source.user.name to declare at least one schema validator (e.g. ConflictsWith id)`)
	}
}

func TestInternetFwRule_ExceptionSourceUser_nameAndIdTogether_invalid(t *testing.T) {
	ctx := context.Background()
	var r internetFwRuleResource
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	validators := exceptionSourceUserNameValidators(t, schemaResp.Schema)

	userElem, cfg := mustUserSetConfig(t, ctx,
		types.StringValue("Jeff Jenkinson"),
		types.StringValue("4"),
	)

	resp := validateExceptionUserName(t, ctx, cfg, userElem, types.StringValue("Jeff Jenkinson"), validators)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics when both name and id are set on exception source.user")
	}
	var found bool
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Summary(), "Invalid Attribute Combination") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected Invalid Attribute Combination diagnostic, got: %s", resp.Diagnostics)
	}
}

func TestInternetFwRule_ExceptionSourceUser_nameOnly_valid(t *testing.T) {
	ctx := context.Background()
	var r internetFwRuleResource
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	validators := exceptionSourceUserNameValidators(t, schemaResp.Schema)

	userElem, cfg := mustUserSetConfig(t, ctx,
		types.StringValue("Jeff Jenkinson"),
		types.StringNull(),
	)

	resp := validateExceptionUserName(t, ctx, cfg, userElem, types.StringValue("Jeff Jenkinson"), validators)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error for name-only reference: %s", resp.Diagnostics)
	}
}

func TestInternetFwRule_ExceptionSourceUser_idOnly_valid(t *testing.T) {
	ctx := context.Background()
	var r internetFwRuleResource
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	validators := exceptionSourceUserNameValidators(t, schemaResp.Schema)

	userElem, cfg := mustUserSetConfig(t, ctx,
		types.StringNull(),
		types.StringValue("4"),
	)

	// Validator is on "name"; when name is null, ConflictsWith returns without error.
	resp := validateExceptionUserName(t, ctx, cfg, userElem, types.StringNull(), validators)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error for id-only reference: %s", resp.Diagnostics)
	}
}
