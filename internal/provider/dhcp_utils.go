package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// checkForDhcpRelayGroup validates DHCP relay group configuration and returns the relay group ID
func checkForDhcpRelayGroup(ctx context.Context, client *catoClientData, relayGroupName, relayGroupId string) (string, bool, error) {
	// Add nil checks
	if client == nil {
		return "", false, fmt.Errorf("client is nil")
	}
	if client.catov2 == nil {
		return "", false, fmt.Errorf("catov2 client is nil")
	}

	hasRelayGroupId := relayGroupId != ""
	hasRelayGroupName := relayGroupName != ""

	// Check that exactly one of relay_group_id or relay_group_name is specified
	if !hasRelayGroupId && !hasRelayGroupName {
		return "", false, fmt.Errorf("When dhcp_type is DHCP_RELAY, either relay_group_id or relay_group_name must be specified")
	}

	if hasRelayGroupId && hasRelayGroupName {
		return "", false, fmt.Errorf("when dhcp_type is DHCP_RELAY, specify either relay_group_id or relay_group_name, but not both")
	}

	// Lookup and validate the DHCP relay group exists
	dhcpRelayGroupResult, err := client.catov2.EntityLookupMinimal(ctx, client.AccountId, cato_models.EntityTypeDhcpRelayGroup, nil, nil, nil, nil, nil)
	tflog.Warn(ctx, "checkForDhcpRelayGroup.dhcpRelayGroupResult.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(dhcpRelayGroupResult),
	})
	if err != nil {
		return "", false, fmt.Errorf("catov2 API EntityLookup error: %w", err)
	}

	// Check if the specified relay group exists
	for _, item := range dhcpRelayGroupResult.EntityLookup.Items {
		if hasRelayGroupId {
			if item.Entity.GetID() == relayGroupId {
				return relayGroupId, true, nil
			}
		} else if hasRelayGroupName {
			if namePtr := item.Entity.GetName(); namePtr != nil && *namePtr == relayGroupName {
				// Return the ID when found by name
				return item.Entity.GetID(), true, nil
			}
		}
	}

	// Relay group not found
	if hasRelayGroupId {
		return "", false, fmt.Errorf("DHCP relay group with ID '%s' not found", relayGroupId)
	} else {
		return "", false, fmt.Errorf("DHCP relay group with name '%s' not found", relayGroupName)
	}
}
