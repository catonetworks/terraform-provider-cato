package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// checkForDhcpRelayGroup validates DHCP relay group configuration and returns the relay group ID
func checkForDhcpRelayGroup(
	ctx context.Context,
	client *catoClientData,
	relayGroupName,
	relayGroupID string,
) (groupID string, found bool, err error) {
	// Add nil checks
	if client == nil {
		return "", false, fmt.Errorf("client is nil")
	}
	if client.catov2 == nil {
		return "", false, fmt.Errorf("catov2 client is nil")
	}

	hasRelayGroupID := relayGroupID != ""
	hasRelayGroupName := relayGroupName != ""

	// Check that exactly one of relay_group_id or relay_group_name is specified
	if !hasRelayGroupID && !hasRelayGroupName {
		return "", false, fmt.Errorf("when dhcp_type is DHCP_RELAY, either relay_group_id or relay_group_name must be specified")
	}

	if hasRelayGroupID && hasRelayGroupName {
		return "", false, fmt.Errorf("when dhcp_type is DHCP_RELAY, specify either relay_group_id or relay_group_name, but not both")
	}

	// If relay_group_id is provided, return it directly without lookup
	// This avoids permission issues with EntityLookup and correctly sends the ID to the API
	if hasRelayGroupID {
		tflog.Debug(ctx, "checkForDhcpRelayGroup: using provided relay_group_id directly", map[string]interface{}{
			"relay_group_id": relayGroupID,
		})
		return relayGroupID, true, nil
	}

	// Only do EntityLookup when relay_group_name is provided (to resolve name to ID)
	dhcpRelayGroupResult, err := client.catov2.EntityLookupMinimal(ctx, client.AccountId, cato_models.EntityTypeDhcpRelayGroup,
		nil, nil, nil, nil, nil)
	tflog.Warn(ctx, "checkForDhcpRelayGroup.dhcpRelayGroupResult.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(dhcpRelayGroupResult),
	})
	if err != nil {
		return "", false, fmt.Errorf("catov2 API EntityLookup error: %w", err)
	}

	// Lookup relay group by name and return the ID
	for _, item := range dhcpRelayGroupResult.EntityLookup.Items {
		if namePtr := item.Entity.GetName(); namePtr != nil && *namePtr == relayGroupName {
			// Return the ID when found by name
			return item.Entity.GetID(), true, nil
		}
	}

	// Relay group not found by name
	return "", false, fmt.Errorf("DHCP relay group with name '%s' not found", relayGroupName)
}
