package provider

import (
	"fmt"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
)

func contains(nameToIdMap map[string]struct{}, name string) bool {
	_, exists := nameToIdMap[name]
	return exists
}

func extractValue(entityType string, item *cato_go_sdk.EntityLookup_EntityLookup_Items) (string, error) {
	switch entityType {
	case "allocatedIP":
		return extractIpAddress(item)
	default:
		return *item.Entity.Name, nil
	}
}

func extractIpAddress(item *cato_go_sdk.EntityLookup_EntityLookup_Items) (string, error) {
	allocatedIp, ok := item.HelperFields["allocatedIp"].(string)
	if !ok {
		return "", fmt.Errorf("failed to read allocatedIp from the helper field of entity %s", *item.Entity.Name)
	} else {
		return allocatedIp, nil
	}
}
