package policy

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
)

type (
	internetFirewallDiscardError = cato_go_sdk.PolicyInternetFirewallDiscardPolicyRevision_Policy_InternetFirewall_DiscardPolicyRevision_Errors
	wanFirewallDiscardError      = cato_go_sdk.PolicyWanFirewallDiscardPolicyRevision_Policy_WanFirewall_DiscardPolicyRevision_Errors
	wanNetworkDiscardError       = cato_go_sdk.PolicyWanNetworkDiscardPolicyRevision_Policy_WanNetwork_DiscardPolicyRevision_Errors
)
