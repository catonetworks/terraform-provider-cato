package entities

type PolicyRevision string

const PublicPolicyRevision PolicyRevision = "PUBLIC"

// PrivateAccessPolicy contains only the fields needed by the current use cases.
type PrivateAccessPolicy struct {
	Rules []PrivateAccessPolicyRule
}

type PrivateAccessPolicyRule struct {
	ID string
}

func (p PrivateAccessPolicy) HasRule(ruleID string) bool {
	for _, rule := range p.Rules {
		if rule.ID == ruleID {
			return true
		}
	}
	return false
}
