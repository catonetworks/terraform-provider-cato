package validators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidIPRange(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"192.0.2.1", true}, {"2001:db8::1", true}, {"192.0.2.0/24", true}, {"2001:db8::/64", true},
		{"192.0.2.1-192.0.2.10", true}, {" 2001:db8::1 - 2001:db8::2 ", true}, {"192.0.2.1-192.0.2.1", true},
		{"192.0.2.10-192.0.2.1", false}, {"2001:db8::2-2001:db8::1", false},
		{"192.0.2.1-2001:db8::1", false}, {"192.0.2.1-garbage", false}, {"garbage-192.0.2.1", false},
		{"192.0.2.1-192.0.2.2-192.0.2.3", false}, {"192.0.2.0/33", false}, {"", false}, {"garbage", false},
	} {
		t.Run(tc.value, func(t *testing.T) { require.Equal(t, tc.valid, IsValidIPRange(tc.value)) })
	}
}
