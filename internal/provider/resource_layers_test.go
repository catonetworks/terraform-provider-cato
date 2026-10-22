package provider

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
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
			for _, filename := range files {
				if strings.HasSuffix(filename, "_test.go") || strings.HasPrefix(filepath.Base(filename), "client_") {
					continue
				}
				require.False(t, strings.HasPrefix(filepath.Base(filename), "hydrate_"), "move hydration into converters or resource orchestration: %s", filename)
				file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
				require.NoError(t, err)
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
						require.NotEqual(t, "Catov2", selector.Sel.Name, "SDK calls belong in client_%s.go: %s", name, filename)
						require.NotEqual(t, "V2", selector.Sel.Name, "SDK access belongs in client_%s.go: %s", name, filename)
						if strings.HasPrefix(filepath.Base(filename), "converters_") {
							require.NotEqual(t, "resourceAPI", selector.Sel.Name, "converters should not fetch or mutate remote objects: %s", filename)
						}
						return true
					})
					return true
				})
			}
		})
	}
}
