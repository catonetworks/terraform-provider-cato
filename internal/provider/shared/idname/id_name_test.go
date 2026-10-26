package idname

import (
	"testing"
)

func TestSchemaNameID(t *testing.T) {
	t.Run("without prefix", func(t *testing.T) {
		s := SchemaNameID("")

		if len(s) != 2 {
			t.Fatalf("expected 2 attributes, got %d", len(s))
		}
		if got := s["name"].GetDescription(); got != "name" {
			t.Fatalf("expected name description %q, got %q", "name", got)
		}
		if got := s["id"].GetDescription(); got != "ID" {
			t.Fatalf("expected id description %q, got %q", "ID", got)
		}
	})

	t.Run("with prefix", func(t *testing.T) {
		s := SchemaNameID("group")

		if got := s["name"].GetDescription(); got != "group name" {
			t.Fatalf("expected name description %q, got %q", "group name", got)
		}
		if got := s["id"].GetDescription(); got != "group ID" {
			t.Fatalf("expected id description %q, got %q", "group ID", got)
		}
	})
}
