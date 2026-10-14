package rbacassignmentapi

// RBACReference always identifies an object by its stable API ID.
type RBACReference struct {
	By    string `json:"by"`
	Input string `json:"input"`
}

type RBACEntityInput struct {
	Type   string        `json:"type"`
	Ref    RBACReference `json:"ref"`
	Action string        `json:"action"`
}

// RBACAssignmentInput is the full replacement of an assignment's roles and scope.
type RBACAssignmentInput struct {
	AssignmentID                 string            `json:"rbacAssignmentId,omitempty"`
	Users                        []RBACReference   `json:"users,omitempty"`
	Groups                       []RBACReference   `json:"groups,omitempty"`
	Roles                        []RBACReference   `json:"roles"`
	AllowedEntities              []RBACEntityInput `json:"allowedEntities"`
	AllowAllSites                string            `json:"allowAllSites"`
	AllowAllUsers                string            `json:"allowAllUsers"`
	AllowViewingUnobfuscatedData *bool             `json:"allowViewingUnobfuscatedData,omitempty"`
	Description                  string            `json:"description"`
}

type RBACObject struct {
	ID string `json:"id" graphql:"id"`
}
type RBACEntity struct {
	ID     string `json:"id" graphql:"id"`
	Type   string `json:"type" graphql:"type"`
	Action string `json:"action" graphql:"action"`
}

type RBACAssignmentAPI struct {
	ID                    string        `json:"id" graphql:"id"`
	Account               *RBACObject   `json:"account" graphql:"account"`
	User                  *RBACObject   `json:"user" graphql:"user"`
	Group                 *RBACObject   `json:"group" graphql:"group"`
	AssignmentType        string        `json:"assignmentType" graphql:"assignmentType"`
	Roles                 []*RBACObject `json:"roles" graphql:"roles"`
	AllowedEntities       []*RBACEntity `json:"allowedEntities" graphql:"allowedEntities"`
	AllowAllSites         string        `json:"allowAllSites" graphql:"allowAllSites"`
	AllowAllUsers         string        `json:"allowAllUsers" graphql:"allowAllUsers"`
	DataObfuscationStatus string        `json:"dataObfuscationStatus" graphql:"dataObfuscationStatus"`
	Description           *string       `json:"description" graphql:"description"`
}

type RBACAssignmentPage struct {
	Items  []*RBACAssignmentAPI `json:"items" graphql:"items"`
	Paging *struct {
		Total int `json:"total" graphql:"total"`
	} `json:"paging" graphql:"paging"`
}
