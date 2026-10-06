package sktsite

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestLayerImports prevents Terraform and SDK types from leaking back into use cases.
func TestLayerImports(t *testing.T) {
	t.Parallel()
	for _, layer := range []string{"application", "adapter"} {
		files, err := filepath.Glob(filepath.Join(layer, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range files {
			if strings.HasSuffix(name, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range file.Imports {
				imported, err := strconv.Unquote(item.Path.Value)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(imported, "terraform-plugin-") || layer == "application" && strings.Contains(imported, "cato-go-sdk") {
					t.Errorf("%s imports forbidden dependency %s", name, imported)
				}
				if layer == "application" && strings.HasSuffix(imported, "/adapter") {
					t.Errorf("%s imports its adapter", name)
				}
			}
		}
	}
}
