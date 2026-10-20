package policy

// BulkPolicySectionRef is one section row from Policy*FirewallSectionsIndex.
type BulkPolicySectionRef struct {
	ID   string
	Name string
}

// BulkPolicyRuleRow is one rule row from Policy*FirewallRulesIndex.
type BulkPolicyRuleRow struct {
	SectionID   string
	SectionName string
	RuleID      string
	RuleName    string
	Index       int64
	IsSystem    bool
}

// BulkPlannedRuleIndex is one rule placement from Terraform rule_data.
type BulkPlannedRuleIndex struct {
	SectionName    string
	RuleName       string
	IndexInSection int64
}
