package validators

import (
	"bytes"
	"net"
	"strings"
)

func IsValidIPRange(value string) bool {
	if net.ParseIP(value) != nil {
		return true
	}
	if _, _, err := net.ParseCIDR(value); err == nil {
		return true
	}

	start, end, ok := strings.Cut(value, "-")
	if !ok || strings.Contains(end, "-") {
		return false
	}

	startIP := net.ParseIP(strings.TrimSpace(start))
	endIP := net.ParseIP(strings.TrimSpace(end))
	if startIP == nil || endIP == nil {
		return false
	}

	// Do not compare IPv4 and IPv6 addresses as one ordered range.
	if (startIP.To4() == nil) != (endIP.To4() == nil) {
		return false
	}

	return bytes.Compare(startIP.To16(), endIP.To16()) <= 0
}
