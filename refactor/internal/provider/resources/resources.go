package resources

import (
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/adapters/catoapi"
	"github.com/catonetworks/terraform-provider-cato/refactor/internal/adapters/memory"
)

// Dependencies holds a configured SDK wrapper and shared infrastructure.
// Terraform lifecycle methods construct their use cases at runtime.
type Dependencies struct {
	CatoAPI      *catoapi.Adapter
	AccountID    string
	ResourceLock *memory.ResourceLock
}
