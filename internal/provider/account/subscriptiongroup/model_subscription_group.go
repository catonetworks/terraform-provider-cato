package subscriptiongroup

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Integration types.Set    `tfsdk:"integration"`
	MailingList types.Set    `tfsdk:"mailing_list"`
}
