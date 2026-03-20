package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VolatileInt64Modifier is a plan modifier for computed fields that can change
// externally (like rule index which changes when other rules are reordered).
// During Update operations, it marks the value as Unknown so Terraform expects
// it to be recomputed after apply, preventing "inconsistent result after apply" errors.
type volatileInt64Modifier struct{}

// VolatileInt64 returns a plan modifier that marks the value as Unknown during
// Update operations. This is used for computed fields like rule index that can
// change due to external factors (other rules being added/removed/reordered).
func VolatileInt64() planmodifier.Int64 {
	return volatileInt64Modifier{}
}

func (m volatileInt64Modifier) Description(_ context.Context) string {
	return "Marks the value as unknown during Update to allow for external changes"
}

func (m volatileInt64Modifier) MarkdownDescription(_ context.Context) string {
	return "Marks the value as unknown during Update to allow for external changes"
}

func (m volatileInt64Modifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// On Create (no prior state), use default behavior - value will be unknown until API returns it
	if req.State.Raw.IsNull() {
		return
	}

	// On Update (prior state exists), mark as unknown because the index can change
	// due to external factors like other rules being added/removed/reordered.
	// This tells Terraform to expect a potentially different value after apply.
	resp.PlanValue = types.Int64Unknown()
}
