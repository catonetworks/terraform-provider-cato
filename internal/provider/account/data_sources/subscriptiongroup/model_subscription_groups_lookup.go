package subscriptiongroup

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Lookup is the type for the subscription groups data source with filters
type Lookup struct {
	IDFilter   types.List `tfsdk:"id_filter"`
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

// MailingListAttrTypes defines the attribute types for a referenced mailing list entry
var MailingListAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

// MailingListObjectType is the object type for a referenced mailing list entry
var MailingListObjectType = types.ObjectType{AttrTypes: MailingListAttrTypes}

// ItemAttrTypes defines the attribute types for a subscription group item in the lookup results
var ItemAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"mailing_list": types.ListType{ElemType: MailingListObjectType},
}
