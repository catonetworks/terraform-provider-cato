package provider

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

// Keep every registered resource on the same mockable API boundary.
func TestRegisteredResourceLayers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := &catoProvider{}
	for _, constructor := range p.Resources(ctx) {
		r := constructor()
		metadata := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, metadata)
		t.Run(metadata.TypeName, func(t *testing.T) {
			t.Parallel()
			typ := reflect.TypeOf(r).Elem()
			packageDir := strings.TrimPrefix(typ.PkgPath(), "github.com/catonetworks/terraform-provider-cato/internal/provider/")
			name := strings.TrimPrefix(metadata.TypeName, "cato_")
			for _, prefix := range []string{"client_", "converters_"} {
				info, err := os.Stat(filepath.Join(packageDir, prefix+name+".go"))
				require.NoError(t, err)
				require.False(t, info.IsDir())
			}
			field, ok := typ.FieldByName("apiClient")
			require.True(t, ok, "resource needs an injectable client")
			require.Equal(t, reflect.Interface, field.Type.Kind())
			require.Positive(t, field.Type.NumMethod(), "client should expose the resource API operations")
			files, err := filepath.Glob(filepath.Join(packageDir, "*.go"))
			require.NoError(t, err)
			packageFiles := make(map[string]*ast.File)
			for _, filename := range files {
				if strings.HasSuffix(filename, "_test.go") {
					continue
				}
				require.False(t, strings.HasPrefix(filepath.Base(filename), "hydrate_"), "move hydration into converters or resource orchestration: %s", filename)
				file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
				require.NoError(t, err)
				packageFiles[filename] = file
			}
			syntax := make([]*ast.File, 0, len(packageFiles))
			for _, file := range packageFiles {
				syntax = append(syntax, file)
			}
			for filename, file := range packageFiles {
				for _, violation := range resourceLayerViolations(filepath.Base(filename), file, syntax...) {
					t.Errorf("%s: %s", filename, violation)
				}
			}
		})
	}
}

// Check responsibilities independently of function names or import aliases.
func resourceLayerViolations(filename string, file *ast.File, packageFiles ...*ast.File) []string {
	client := strings.HasPrefix(filename, "client_")
	converter := strings.HasPrefix(filename, "converters_")
	aliases, violations := resourceLayerImports(file, client, converter)
	apiMethods := resourceAPIMethods(append(packageFiles, file))
	if converter {
		violations = append(violations, converterDiagnosticViolations(file, aliases)...)
	}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ast.Inspect(call.Fun, func(n ast.Node) bool {
			selector, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if !client && (selector.Sel.Name == "Catov2" || selector.Sel.Name == "V2") {
				violations = append(violations, "SDK access belongs in the client file")
			}
			if converter && (selector.Sel.Name == "resourceAPI" || (call.Fun == selector && apiMethods[selector.Sel.Name])) {
				violations = append(violations, "converters must not fetch or mutate remote objects")
			}
			if !client {
				return true
			}
			if qualifier, ok := selector.X.(*ast.Ident); ok {
				switch aliases[qualifier.Name] {
				case "github.com/hashicorp/terraform-plugin-framework/types",
					"github.com/hashicorp/terraform-plugin-framework/types/basetypes":
					violations = append(violations, "Terraform value construction belongs in converters")
				}
			}
			if clientDecodesTerraformValue(selector, aliases) {
				violations = append(violations, "Terraform value decoding belongs outside the client file")
			}

			return true
		})
		return true
	})
	return violations
}

func clientDecodesTerraformValue(selector *ast.SelectorExpr, aliases map[string]string) bool {
	if qualifier, ok := selector.X.(*ast.Ident); ok && aliases[qualifier.Name] == "errors" && selector.Sel.Name == "As" {
		return false
	}
	switch selector.Sel.Name {
	case "As", "ElementsAs", "ValueString", "ValueStringPointer", "ValueBool", "ValueBoolPointer",
		"ValueInt64", "ValueInt64Pointer", "ValueFloat64", "ValueFloat64Pointer", "IsNull", "IsUnknown":
		return true
	}
	return false
}

func resourceLayerImports(file *ast.File, client, converter bool) (aliases map[string]string, violations []string) {
	aliases = make(map[string]string)
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		alias := filepath.Base(importPath)
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		aliases[alias] = importPath
		if (client || converter) && importPath == "testing" {
			violations = append(violations, "test helpers belong outside client and converter files")
		}
		if client && importPath == "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel" {
			violations = append(violations, "client interfaces should use SDK inputs and responses rather than Terraform models")
		}
	}
	return aliases, violations
}

func TestResourceLayerViolations(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, filename, source string
		wantViolation          bool
	}{
		{"SDK response conversion despite Client suffix", "client_rule.go", `package example
            import values "github.com/hashicorp/terraform-plugin-framework/types"
            func hydrateFromClient() { _ = values.StringValue("rule") }`, true},
		{"standard error inspection", "client_rule.go", `package example
            import inspect "errors"
            func unavailable(err error) bool { var response *APIError; return inspect.As(err, &response) }`, false},
		{"Terraform object decoding", "client_rule.go", `package example
            func decode(value Object) { value.As(ctx, &model, options) }`, true},
		{"inline Terraform decoding", "client_rule.go", `package example
            func update(plan Model) { _ = plan.Name.ValueString() }`, true},
		{"SDK forwarding", "client_rule.go", `package example
            func update() { sdk.Update(ctx, input) }`, false},
		{"Terraform model in client", "client_rule.go", `package example
            import tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
            func update() tf.GlobalIPRange { return tf.GlobalIPRange{} }`, true},
		{"conversion", "converters_rule.go", `package example
            import values "github.com/hashicorp/terraform-plugin-framework/types"
            func hydrateFromClient() { _ = values.StringValue("rule") }`, false},
		{"API in converter", "converters_rule.go", `package example
            func hydrate() { r.resourceAPI().Read(ctx) }`, true},
		{"API under another accessor", "converters_rule.go", `package example
            type ResourceClient interface { Fetch() }
            func hydrate() { r.anotherAccessor().Fetch() }`, true},
		{"SDK response field is not an API call", "converters_rule.go", `package example
            type ResourceClient interface { AccountSnapshot() }
            func hydrate(response Snapshot) { response.AccountSnapshot.GetSites() }`, false},
		{"lost diagnostics", "converters_rule.go", `package example
            import report "github.com/hashicorp/terraform-plugin-framework/diag"
            func encode(d report.Diagnostics) { d.Append(errs...) }`, true},
		{"returned diagnostics", "converters_rule.go", `package example
            import report "github.com/hashicorp/terraform-plugin-framework/diag"
            func encode(d report.Diagnostics) report.Diagnostics { d.Append(errs...); return d }`, false},
		{"SDK access in resource", "resource_rule.go", `package example
            func read() { r.client.Catov2.Read(ctx) }`, true},
		{"test helpers in converter", "converters_rule.go", `package example
            import "testing"
            func schema(t *testing.T) { t.Helper() }`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			file, err := parser.ParseFile(token.NewFileSet(), tc.filename, tc.source, 0)
			require.NoError(t, err)
			require.Equal(t, tc.wantViolation, len(resourceLayerViolations(tc.filename, file)) > 0)
		})
	}
}

// Resolve operation names from the package interface rather than accessor spelling.
func resourceAPIMethods(files []*ast.File) map[string]bool {
	methods := make(map[string]bool)
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "ResourceClient" {
				return true
			}
			iface, ok := spec.Type.(*ast.InterfaceType)
			if !ok {
				return false
			}
			for _, field := range iface.Methods.List {
				for _, name := range field.Names {
					methods[name.Name] = true
				}
			}
			return false
		})
	}
	return methods
}

// Diagnostics are slices: a value parameter cannot propagate appended entries.
func converterDiagnosticViolations(file *ast.File, aliases map[string]string) []string {
	isDiagnostics := func(expr ast.Expr) bool {
		selector, ok := expr.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Diagnostics" {
			return false
		}
		qualifier, ok := selector.X.(*ast.Ident)
		return ok && aliases[qualifier.Name] == "github.com/hashicorp/terraform-plugin-framework/diag"
	}
	var violations []string
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		returnsDiagnostics := false
		if function.Type.Results != nil {
			for _, field := range function.Type.Results.List {
				returnsDiagnostics = returnsDiagnostics || isDiagnostics(field.Type)
			}
		}
		if returnsDiagnostics {
			continue
		}
		for _, field := range function.Type.Params.List {
			if isDiagnostics(field.Type) {
				violations = append(violations, function.Name.Name+": pass diagnostics by pointer or return them to the caller")
			}
		}
	}
	return violations
}
