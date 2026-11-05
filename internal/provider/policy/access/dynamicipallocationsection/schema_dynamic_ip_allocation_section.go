package dynamicipallocationsection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a dynamic ip allocation section. Changes are automatically " +
			"published. Only empty sections " +
			"can be deleted.",
		Attributes: policyshared.SectionResourceAttributes(),
	}
}
