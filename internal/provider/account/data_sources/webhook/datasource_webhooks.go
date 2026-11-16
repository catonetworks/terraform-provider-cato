package webhook

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func NewDataSource() datasource.DataSource {
	return &webhooksDataSource{}
}

type webhooksDataSource struct {
	client *catoclient.Service
}

func (d *webhooksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhooks"
}

func (d *webhooksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_webhooks` data source fetches information about account webhook integrations " +
			"with optional filters. Returns a list of webhooks matching the filter criteria. Never exposes " +
			"`auth` or `body`: the underlying list query never selects webhook secrets.",
		Attributes: map[string]schema.Attribute{
			"id_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of webhook IDs to filter by",
				Optional:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"name_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of webhook names to filter by",
				Optional:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"items": schema.ListNestedAttribute{
				Description: "List of webhooks matching the filter criteria",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                 schema.StringAttribute{Description: "Webhook ID", Computed: true},
						"name":               schema.StringAttribute{Description: "Webhook name", Computed: true},
						"enabled":            schema.BoolAttribute{Description: "Whether the webhook is enabled", Computed: true},
						"url":                schema.StringAttribute{Description: "Destination URL", Computed: true},
						"request_method":     schema.StringAttribute{Description: "HTTP request method", Computed: true},
						"related_webhook_id": schema.StringAttribute{Description: "Related webhook ID, if any", Computed: true},
						"header": schema.ListNestedAttribute{
							Description: "Additional HTTP headers sent with the request", Computed: true,
							NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
								"name":  schema.StringAttribute{Computed: true},
								"value": schema.StringAttribute{Computed: true, Sensitive: true},
							}},
						},
						"response_mapping": schema.ListNestedAttribute{
							Description: "Key/value pairs mapped out of the webhook response", Computed: true,
							NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
								"key":   schema.StringAttribute{Computed: true},
								"value": schema.StringAttribute{Computed: true},
							}},
						},
					},
				},
			},
		},
	}
}

func (d *webhooksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*catoclient.Service)
}

// webhookListDTO/webhookListResponse are hand-written (not the generated
// Catov2.NotificationIntegrationWebhookList method) because the CLI-sourced field
// selection the generated query uses -- and therefore the generated result type's shape
// -- requests `auth { ... on CustomHeadersAuth { ... } }`, a GraphQL union member not yet
// deployed on every tenant's notification microservice; selecting it fails at runtime with
// a DOWNSTREAM_SERVICE_ERROR / UnknownType validation error. This data source never
// exposes secrets anyway, so it selects only the non-secret fields below and skips `auth`
// (and `body`) entirely. See internal/provider/account/webhook/client_webhook.go for the
// matching deviation on the resource side.
type webhookListDTO struct {
	ID               string                      `json:"id"`
	Name             string                      `json:"name"`
	Enabled          bool                        `json:"enabled"`
	URL              string                      `json:"url"`
	RequestMethod    string                      `json:"requestMethod"`
	Header           []webhookListHeaderDTO      `json:"header"`
	ResponseMapping  []webhookListResponseMapDTO `json:"responseMapping"`
	RelatedWebhookID *string                     `json:"relatedWebhookId"`
}
type webhookListHeaderDTO struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type webhookListResponseMapDTO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type webhookListResponse struct {
	Notification struct {
		Integration struct {
			WebhookList struct {
				Items []webhookListDTO `json:"items"`
			} `json:"webhookList"`
		} `json:"integration"`
	} `json:"notification"`
}

const webhookListDocument = `query notificationIntegrationWebhookList($accountId: ID!) {
notification(accountId: $accountId) {
integration {
webhookList {
items {
id
name
enabled
url
requestMethod
header { name value }
responseMapping { key value }
relatedWebhookId
}
}
}
}
}`

//nolint:gocyclo,funlen
func (d *webhooksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Lookup
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result := new(webhookListResponse)
	err := d.client.V2().Client.Post(ctx, "notificationIntegrationWebhookList", webhookListDocument, result,
		map[string]interface{}{"accountId": d.client.AccountId})
	tflog.Debug(ctx, "DataSource.notificationIntegrationWebhookList.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		resp.Diagnostics.AddError("Cato API notificationIntegrationWebhookList error", err.Error())
		return
	}

	filterByID := !config.IDFilter.IsNull() && config.IDFilter.Elements() != nil
	idsMap := make(map[string]struct{})
	if filterByID {
		for _, value := range config.IDFilter.Elements() {
			idsMap[strings.Trim(value.String(), "\"")] = struct{}{}
		}
	}
	filterByName := !config.NameFilter.IsNull() && config.NameFilter.Elements() != nil
	namesMap := make(map[string]struct{})
	if filterByName {
		for _, value := range config.NameFilter.Elements() {
			namesMap[strings.Trim(value.String(), "\"")] = struct{}{}
		}
	}

	var listObjects []attr.Value
	for _, item := range result.Notification.Integration.WebhookList.Items {
		var shouldInclude bool
		switch {
		case filterByID && filterByName:
			shouldInclude = utils.Contains(idsMap, item.ID) && utils.Contains(namesMap, item.Name)
		case filterByID:
			shouldInclude = utils.Contains(idsMap, item.ID)
		case filterByName:
			shouldInclude = utils.Contains(namesMap, item.Name)
		default:
			shouldInclude = true
		}
		if !shouldInclude {
			continue
		}

		headerValues := make([]attr.Value, 0, len(item.Header))
		for _, h := range item.Header {
			obj, diags := types.ObjectValue(HeaderAttrTypes, map[string]attr.Value{
				"name": types.StringValue(h.Name), "value": types.StringValue(h.Value),
			})
			resp.Diagnostics.Append(diags...)
			headerValues = append(headerValues, obj)
		}
		headerList, diags := types.ListValue(HeaderObjectType, headerValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		responseMapValues := make([]attr.Value, 0, len(item.ResponseMapping))
		for _, rm := range item.ResponseMapping {
			obj, diags := types.ObjectValue(ResponseMapAttrTypes, map[string]attr.Value{
				"key": types.StringValue(rm.Key), "value": types.StringValue(rm.Value),
			})
			resp.Diagnostics.Append(diags...)
			responseMapValues = append(responseMapValues, obj)
		}
		responseMapList, diags := types.ListValue(ResponseMapObjectType, responseMapValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		relatedID := types.StringNull()
		if item.RelatedWebhookID != nil {
			relatedID = types.StringValue(*item.RelatedWebhookID)
		}

		listObj, diags := types.ObjectValue(ItemAttrTypes, map[string]attr.Value{
			"id":                 types.StringValue(item.ID),
			"name":               types.StringValue(item.Name),
			"enabled":            types.BoolValue(item.Enabled),
			"url":                types.StringValue(item.URL),
			"request_method":     types.StringValue(item.RequestMethod),
			"header":             headerList,
			"response_mapping":   responseMapList,
			"related_webhook_id": relatedID,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		listObjects = append(listObjects, listObj)
	}

	itemsList, diags := types.ListValue(types.ObjectType{AttrTypes: ItemAttrTypes}, listObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Items = itemsList
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
