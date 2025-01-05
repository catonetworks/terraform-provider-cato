package provider

import (
	"strings"
)

func contains(nameToIdMap map[string]struct{}, name string) bool {
	_, exists := nameToIdMap[name]
	return exists
}

func extractValue(entityType string, input string) string {
	switch entityType {
	case "allocatedIP":
		return extractIpAddress(input)
	default:
		return input
	}
}

func extractIpAddress(input string) string {
	index := strings.Index(input, " - ")
	if index != -1 {
		return input[index+3:]
	}
	return ""
}
