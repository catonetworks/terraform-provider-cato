package webhook

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Lookup is the type for the webhooks data source with filters.
type Lookup struct {
	IDFilter   types.List `tfsdk:"id_filter"`
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

// HeaderAttrTypes/HeaderObjectType describe one `header` list element.
var HeaderAttrTypes = map[string]attr.Type{"name": types.StringType, "value": types.StringType}
var HeaderObjectType = types.ObjectType{AttrTypes: HeaderAttrTypes}

// ResponseMapAttrTypes/ResponseMapObjectType describe one `response_mapping` list element.
var ResponseMapAttrTypes = map[string]attr.Type{"key": types.StringType, "value": types.StringType}
var ResponseMapObjectType = types.ObjectType{AttrTypes: ResponseMapAttrTypes}

// ItemAttrTypes defines the attribute types for a webhook item in the lookup results.
// Secrets are never exposed here: the underlying list query never selects `auth` or
// `body` (see datasource_webhooks.go).
var ItemAttrTypes = map[string]attr.Type{
	"id":                 types.StringType,
	"name":               types.StringType,
	"enabled":            types.BoolType,
	"url":                types.StringType,
	"request_method":     types.StringType,
	"header":             types.ListType{ElemType: HeaderObjectType},
	"response_mapping":   types.ListType{ElemType: ResponseMapObjectType},
	"related_webhook_id": types.StringType,
}
