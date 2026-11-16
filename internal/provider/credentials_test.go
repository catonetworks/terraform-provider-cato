package provider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCredsFile(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write fixture credentials file: %v", err)
	}
	return path
}

func TestParseINISections(t *testing.T) {
	t.Parallel()

	input := `
# a comment
; another comment style

[refarch-dev]
endpoint = https://api.catonetworks.com/api/v1/graphql2
cato_token = abc123
account_id = 15412
scim_url = https://scim.example.com/%2Fv2
cookie = session=abc; Path=/

[prod]
cato_token: def456
account_id: 99999
`

	sections, err := parseINISections(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d: %v", len(sections), sections)
	}

	refarch, ok := sections["refarch-dev"]
	if !ok {
		t.Fatalf("expected section %q to be present", "refarch-dev")
	}
	if got, want := refarch["cato_token"], "abc123"; got != want {
		t.Errorf("cato_token = %q, want %q", got, want)
	}
	if got, want := refarch["account_id"], "15412"; got != want {
		t.Errorf("account_id = %q, want %q", got, want)
	}
	// Values containing literal "%" must survive untouched (no interpolation).
	if got, want := refarch["scim_url"], "https://scim.example.com/%2Fv2"; got != want {
		t.Errorf("scim_url = %q, want %q", got, want)
	}
	// Arbitrary/unknown keys are retained by the parser (filtering happens in
	// ResolveProfileCredentials, not here).
	if got, want := refarch["cookie"], "session=abc; Path=/"; got != want {
		t.Errorf("cookie = %q, want %q", got, want)
	}

	prod, ok := sections["prod"]
	if !ok {
		t.Fatalf("expected section %q to be present", "prod")
	}
	if got, want := prod["cato_token"], "def456"; got != want {
		t.Errorf("cato_token = %q, want %q", got, want)
	}
	if got, want := prod["account_id"], "99999"; got != want {
		t.Errorf("account_id = %q, want %q", got, want)
	}
}

func TestParseINISectionsIgnoresKeysBeforeFirstSection(t *testing.T) {
	t.Parallel()

	input := "cato_token = orphaned\n[profile]\ncato_token = real\naccount_id = 1\n"

	sections, err := parseINISections(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := sections[""]; ok {
		t.Errorf("expected no section for keys preceding any header")
	}
	if got, want := sections["profile"]["cato_token"], "real"; got != want {
		t.Errorf("cato_token = %q, want %q", got, want)
	}
}

func TestSplitINIKeyValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{name: "equals delimiter", line: "cato_token = abc123", wantKey: "cato_token", wantValue: "abc123", wantOK: true},
		{name: "colon delimiter", line: "cato_token: abc123", wantKey: "cato_token", wantValue: "abc123", wantOK: true},
		{name: "value containing percent", line: "scim_url = https://x/%2F", wantKey: "scim_url", wantValue: "https://x/%2F", wantOK: true},
		{name: "value containing equals", line: "cookie = a=b=c", wantKey: "cookie", wantValue: "a=b=c", wantOK: true},
		{name: "no delimiter", line: "not-a-kv-line", wantOK: false},
		{name: "empty key", line: "= novalue", wantOK: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key, value, ok := splitINIKeyValue(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
			if value != tt.wantValue {
				t.Errorf("value = %q, want %q", value, tt.wantValue)
			}
		})
	}
}

//nolint:gocyclo
func TestResolveProfileCredentials(t *testing.T) {
	t.Parallel()

	fixture := `[refarch-dev]
endpoint = https://api.catonetworks.com/api/v1/graphql2
cato_token = abc123
account_id = 15412
scim_url = https://scim.example.com

[no-token]
account_id = 1

[no-account]
cato_token = abc123
`
	path := writeCredsFile(t, fixture)

	t.Run("resolves token account endpoint", func(t *testing.T) {
		t.Parallel()
		token, accountID, endpoint, err := ResolveProfileCredentials("refarch-dev", path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != "abc123" {
			t.Errorf("token = %q, want %q", token, "abc123")
		}
		if accountID != "15412" {
			t.Errorf("accountID = %q, want %q", accountID, "15412")
		}
		if endpoint != "https://api.catonetworks.com/api/v1/graphql2" {
			t.Errorf("endpoint = %q, want %q", endpoint, "https://api.catonetworks.com/api/v1/graphql2")
		}
	})

	t.Run("missing section lists available profiles, not values", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := ResolveProfileCredentials("does-not-exist", path)
		if err == nil {
			t.Fatal("expected an error for a missing profile section")
		}
		msg := err.Error()
		if !strings.Contains(msg, "no-account") || !strings.Contains(msg, "no-token") || !strings.Contains(msg, "refarch-dev") {
			t.Errorf("error should list available profile names, got: %s", msg)
		}
		if strings.Contains(msg, "abc123") {
			t.Errorf("error must not leak credential values, got: %s", msg)
		}
	})

	t.Run("missing cato_token key errors", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := ResolveProfileCredentials("no-token", path)
		if err == nil || !strings.Contains(err.Error(), "cato_token") {
			t.Fatalf("expected a cato_token-missing error, got: %v", err)
		}
	})

	t.Run("missing account_id key errors", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := ResolveProfileCredentials("no-account", path)
		if err == nil || !strings.Contains(err.Error(), "account_id") {
			t.Fatalf("expected an account_id-missing error, got: %v", err)
		}
	})

	t.Run("missing file errors without leaking path contents", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := ResolveProfileCredentials("refarch-dev", filepath.Join(t.TempDir(), "does-not-exist"))
		if err == nil {
			t.Fatal("expected an error for a missing credentials file")
		}
	})

	t.Run("empty profile name errors", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := ResolveProfileCredentials("", path)
		if err == nil {
			t.Fatal("expected an error for an empty profile name")
		}
	})

	t.Run("empty path errors", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := ResolveProfileCredentials("refarch-dev", "")
		if err == nil {
			t.Fatal("expected an error for an empty credentials file path")
		}
	})
}

func TestCredentialsFilePathHonorsEnvOverride(t *testing.T) {
	// Not parallel: mutates a process-wide environment variable.
	custom := filepath.Join(t.TempDir(), "custom-credentials")
	t.Setenv(credentialsFileEnvVar, custom)

	if got := credentialsFilePath(); got != custom {
		t.Errorf("credentialsFilePath() = %q, want %q", got, custom)
	}
}

func TestCredentialsFilePathDefaultsToDotCato(t *testing.T) {
	// Not parallel: mutates a process-wide environment variable.
	t.Setenv(credentialsFileEnvVar, "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home directory in this environment: %v", err)
	}

	want := filepath.Join(home, ".cato", "credentials")
	if got := credentialsFilePath(); got != want {
		t.Errorf("credentialsFilePath() = %q, want %q", got, want)
	}
}
