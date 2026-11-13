package usersgroup

import (
	"context"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// terraform field names
const (
	tfNameFilter       = "name_filter"
	tfImportTypeFilter = "import_type_filter"
	tfItems            = "items"
	tfID               = "id"
	tfName             = "name"
	tfDescription      = "description"
	tfImportType       = "import_type"
	tfMembersCount     = "members_count"
)

// pageLimit is the page size used when paginating usersGroupList to exhaustion.
const pageLimit = 1000

var attrTypes = map[string]attr.Type{
	tfID:           types.StringType,
	tfName:         types.StringType,
	tfDescription:  types.StringType,
	tfImportType:   types.StringType,
	tfMembersCount: types.Int64Type,
}

func DataSource() datasource.DataSource {
	return &usersGroupDataSource{}
}

type usersGroupDataSource struct {
	client *catoclient.Service
}

func (d *usersGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users_group"
}

func (d *usersGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves users groups.",
		Attributes: map[string]schema.Attribute{
			tfNameFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of group names to filter",
				Optional:    true,
			},
			tfImportTypeFilter: schema.StringAttribute{
				Description: "Filter by group import type (DYNAMIC, INTEGRATION, LDAP, MANUAL, SCIM)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("DYNAMIC", "INTEGRATION", "LDAP", "MANUAL", "SCIM"),
				},
			},

			tfItems: schema.ListNestedAttribute{
				Description: "List of users groups",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						tfID: schema.StringAttribute{
							Description: "ID",
							Computed:    true,
						},
						tfName: schema.StringAttribute{
							Description: "Name",
							Computed:    true,
						},
						tfDescription: schema.StringAttribute{
							Description: "Description",
							Computed:    true,
						},
						tfImportType: schema.StringAttribute{
							Description: "Import type",
							Computed:    true,
						},
						tfMembersCount: schema.Int64Attribute{
							Description: "Number of members in the group",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *usersGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

func (d *usersGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var groupLookup Lookup
	if diags := req.Config.Get(ctx, &groupLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	apiFilter := importTypeFilter(groupLookup.ImportTypeFilter)

	var allItems []*cato_go_sdk.UserUsersGroupList_User_UsersGroupList_Items
	from := int64(0)
	for {
		usersGroupListInput := cato_models.UsersGroupListInput{
			Filter: apiFilter,
			Paging: &cato_models.PagingInput{
				From:  from,
				Limit: pageLimit,
			},
		}

		result, err := d.client.Catov2.UserUsersGroupList(ctx, d.client.AccountId, usersGroupListInput)
		tflog.Debug(ctx, "Read.UserUsersGroupList.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API UserUsersGroupList error", err.Error())
			return
		}

		usersGroupList := result.GetUser().GetUsersGroupList()
		if usersGroupList == nil {
			break
		}

		items := usersGroupList.GetItems()
		allItems = append(allItems, items...)

		total := int64(0)
		if paging := usersGroupList.GetPaging(); paging != nil {
			total = paging.GetTotal()
		}
		from += int64(len(items))
		if len(items) == 0 || from >= total {
			break
		}
	}

	filter := makeFilter(&groupLookup)
	var objects []attr.Value

	for _, item := range allItems {
		name := item.GetName()
		if !filter.accept(name) {
			continue
		}

		description := cast.ToString(item.GetDescription())
		importType := ""
		if it := item.GetImportType(); it != nil {
			importType = string(*it)
		}

		obj, diags := types.ObjectValue(
			attrTypes,
			map[string]attr.Value{
				tfID:           types.StringValue(item.GetID()),
				tfName:         types.StringValue(name),
				tfDescription:  types.StringValue(description),
				tfImportType:   types.StringValue(importType),
				tfMembersCount: types.Int64Value(item.GetMembersCount()),
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		objects = append(objects, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objects)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	groupLookup.Items = list
	if diags := resp.State.Set(ctx, &groupLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

// makeFilter prepares a lookup map for names;
// if IsNull() then the map is nil
func makeFilter(groupLookup *Lookup) *usersGroupFilter {
	var f usersGroupFilter

	if !groupLookup.NameFilter.IsNull() {
		f.namesMap = make(map[string]struct{})
		for _, value := range groupLookup.NameFilter.Elements() {
			valueStr := strings.Trim(value.String(), "\"") // Trim quotes
			f.namesMap[valueStr] = struct{}{}
		}
	}
	return &f
}

// accept returns true if the item should be returned
func (f *usersGroupFilter) accept(name string) bool {
	if f.namesMap == nil {
		return true // no filter applied
	}
	_, ok := f.namesMap[name]
	return ok
}

func importTypeFilter(value types.String) *cato_models.UsersGroupFilterInput {
	var apiFilter *cato_models.UsersGroupFilterInput
	if !value.IsNull() && value.ValueString() != "" {
		apiFilter = &cato_models.UsersGroupFilterInput{
			ImportType: []*cato_models.UsersGroupImportTypeFilterInput{
				{
					In: []cato_models.UsersGroupImportType{
						cato_models.UsersGroupImportType(value.ValueString()),
					},
				},
			},
		}
	}

	return apiFilter
}
