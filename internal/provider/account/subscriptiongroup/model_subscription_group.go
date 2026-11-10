package subscriptiongroup

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	// Integration and MailingList are Lists (not Sets) so that name-only references
	// -- where the Computed `id` is Unknown at plan time -- can be correlated after
	// apply. Terraform matches List elements by index; a Set element containing an
	// Unknown value cannot be correlated and raises "inconsistent result after apply".
	Integration types.List `tfsdk:"integration"`
	MailingList types.List `tfsdk:"mailing_list"`
}
