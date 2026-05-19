package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// checkForDhcpRelayGroup validates DHCP relay group configuration and returns the relay group ID
//
// nolint:gocyclo
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

	// Lookup and validate the DHCP relay group exists
	dhcpRelayGroupResult, err := client.catov2.EntityLookupMinimal(
		ctx, client.AccountId, cato_models.EntityTypeDhcpRelayGroup, nil, nil, nil, nil, nil,
	)
	tflog.Warn(ctx, "checkForDhcpRelayGroup.dhcpRelayGroupResult.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(dhcpRelayGroupResult),
	})
	if err != nil {
		return "", false, fmt.Errorf("catov2 API EntityLookup error: %w", err)
	}

	// Check if the specified relay group exists
	for _, item := range dhcpRelayGroupResult.EntityLookup.Items {
		if hasRelayGroupID {
			if item.Entity.GetID() == relayGroupID {
				name := item.Entity.GetName()
				if name == nil {
					return "", false, fmt.Errorf("failed to get dhcpRelayGroup name for id %s", relayGroupID)
				}
				return *name, true, nil
			}
		} else if hasRelayGroupName {
			if namePtr := item.Entity.GetName(); namePtr != nil && *namePtr == relayGroupName {
				// Return the ID when found by name
				return item.Entity.GetID(), true, nil
			}
		}
	}

	// Relay group not found
	if hasRelayGroupID {
		return "", false, fmt.Errorf("DHCP relay group with ID '%s' not found", relayGroupID)
	}
	return "", false, fmt.Errorf("DHCP relay group with name '%s' not found", relayGroupName)
}
