package provider

import "testing"

func TestNormalizedInterfaceIDCandidates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "INT format",
			input:    "INT_1",
			expected: []string{"INT_1", "1"},
		},
		{
			name:     "WAN format",
			input:    "WAN1",
			expected: []string{"WAN1", "1", "INT_1"},
		},
		{
			name:     "numeric format",
			input:    "2",
			expected: []string{"2", "INT_2", "WAN2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			candidates := normalizedInterfaceIDCandidates(tt.input)
			if len(candidates) == 0 {
				t.Fatalf("expected candidates for %q, got none", tt.input)
			}

			set := make(map[string]struct{}, len(candidates))
			for _, c := range candidates {
				set[c] = struct{}{}
			}

			for _, expected := range tt.expected {
				if _, ok := set[expected]; !ok {
					t.Fatalf("expected candidate %q in %v", expected, candidates)
				}
			}
		})
	}
}

func TestWanInterfaceIDsMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resource  string
		snapshot  string
		device    string
		shouldHit bool
	}{
		{
			name:      "exact INT match",
			resource:  "INT_1",
			snapshot:  "INT_1",
			device:    "1",
			shouldHit: true,
		},
		{
			name:      "WAN and INT aliases match",
			resource:  "WAN1",
			snapshot:  "INT_1",
			device:    "1",
			shouldHit: true,
		},
		{
			name:      "numeric and INT aliases match",
			resource:  "INT_2",
			snapshot:  "2",
			device:    "2",
			shouldHit: true,
		},
		{
			name:      "different interfaces do not match",
			resource:  "INT_1",
			snapshot:  "INT_2",
			device:    "2",
			shouldHit: false,
		},
		{
			name:      "selected numeric interface does not match different device",
			resource:  "3",
			snapshot:  "1",
			device:    "",
			shouldHit: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wanInterfaceIDsMatch(tt.resource, tt.snapshot, tt.device)
			if got != tt.shouldHit {
				t.Fatalf("wanInterfaceIDsMatch(%q, %q, %q) = %v, want %v", tt.resource, tt.snapshot, tt.device, got, tt.shouldHit)
			}
		})
	}
}

func TestWanRoleFromInterfaceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		interfaceID string
		expected    string
	}{
		{
			name:        "bare numeric wan 1",
			interfaceID: "1",
			expected:    "wan_1",
		},
		{
			name:        "bare numeric wan 2",
			interfaceID: "2",
			expected:    "wan_2",
		},
		{
			name:        "INT format",
			interfaceID: "INT_3",
			expected:    "wan_3",
		},
		{
			name:        "WAN format",
			interfaceID: "WAN4",
			expected:    "wan_4",
		},
		{
			name:        "unknown",
			interfaceID: "LAN1",
			expected:    "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := wanRoleFromInterfaceID(tt.interfaceID); got != tt.expected {
				t.Fatalf("wanRoleFromInterfaceID(%q) = %q, want %q", tt.interfaceID, got, tt.expected)
			}
		})
	}
}

func TestPrecedenceFromNaturalOrder(t *testing.T) {
	t.Parallel()

	int64Ptr := func(v int64) *int64 {
		return &v
	}

	tests := []struct {
		name     string
		input    *int64
		expected string
		isNull   bool
	}{
		{
			name:     "active precedence",
			input:    int64Ptr(1),
			expected: "ACTIVE",
		},
		{
			name:     "passive precedence",
			input:    int64Ptr(2),
			expected: "PASSIVE",
		},
		{
			name:     "last resort precedence",
			input:    int64Ptr(3),
			expected: "LAST_RESORT",
		},
		{
			name:   "unknown precedence returns null",
			input:  int64Ptr(99),
			isNull: true,
		},
		{
			name:   "nil precedence returns null",
			input:  nil,
			isNull: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := precedenceFromNaturalOrder(tt.input)
			if tt.isNull {
				if !got.IsNull() {
					t.Fatalf("expected null precedence, got %q", got.ValueString())
				}
				return
			}

			if got.IsNull() || got.IsUnknown() {
				t.Fatalf("expected precedence %q, got null/unknown", tt.expected)
			}
			if got.ValueString() != tt.expected {
				t.Fatalf("expected precedence %q, got %q", tt.expected, got.ValueString())
			}
		})
	}
}
