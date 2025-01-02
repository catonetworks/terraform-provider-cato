package provider

import "strings"

func contains(idMap map[string]struct{}, id string) bool {
	_, exists := idMap[id]
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
