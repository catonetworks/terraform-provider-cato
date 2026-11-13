package users

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
	tfEmailFilter      = "email_filter"
	tfImportTypeFilter = "import_type_filter"
	tfItems            = "items"
	tfID               = "id"
	tfEmail            = "email"
	tfFirstName        = "first_name"
	tfLastName         = "last_name"
	tfUserStatus       = "user_status"
	tfUserImportType   = "user_import_type"
)

// pageLimit is the page size used when paginating userList to exhaustion.
const pageLimit = 1000

var attrTypes = map[string]attr.Type{
	tfID:             types.StringType,
	tfEmail:          types.StringType,
	tfFirstName:      types.StringType,
	tfLastName:       types.StringType,
	tfUserStatus:     types.StringType,
	tfUserImportType: types.StringType,
}

func DataSource() datasource.DataSource {
	return &usersDataSource{}
}

type usersDataSource struct {
	client *catoclient.Service
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves account users.",
		Attributes: map[string]schema.Attribute{
			tfNameFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of user display names to filter",
				Optional:    true,
			},
			tfEmailFilter: schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of user emails to filter",
				Optional:    true,
			},
			tfImportTypeFilter: schema.StringAttribute{
				Description: "Filter by user import type (INTEGRATION, LDAP, MANUAL, SCIM)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("INTEGRATION", "LDAP", "MANUAL", "SCIM"),
				},
			},

			tfItems: schema.ListNestedAttribute{
				Description: "List of users",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						tfID: schema.StringAttribute{
							Description: "ID",
							Computed:    true,
						},
						tfEmail: schema.StringAttribute{
							Description: "Email",
							Computed:    true,
						},
						tfFirstName: schema.StringAttribute{
							Description: "First name",
							Computed:    true,
						},
						tfLastName: schema.StringAttribute{
							Description: "Last name",
							Computed:    true,
						},
						tfUserStatus: schema.StringAttribute{
							Description: "User status",
							Computed:    true,
						},
						tfUserImportType: schema.StringAttribute{
							Description: "User import type",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*catoclient.Service)
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var userLookup Lookup
	if diags := req.Config.Get(ctx, &userLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	apiFilter := importTypeFilter(userLookup.ImportTypeFilter)

	var allItems []*cato_go_sdk.UserUserList_User_UserList_Items
	from := int64(0)
	for {
		userListInput := cato_models.UserListInput{
			Filter: apiFilter,
			Paging: &cato_models.PagingInput{
				From:  from,
				Limit: pageLimit,
			},
		}

		result, err := d.client.Catov2.UserUserList(ctx, d.client.AccountId, userListInput)
		tflog.Debug(ctx, "Read.UserUserList.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(result),
		})
		if err != nil {
			resp.Diagnostics.AddError("Catov2 API UserUserList error", err.Error())
			return
		}

		userList := result.GetUser().GetUserList()
		if userList == nil {
			break
		}

		items := userList.GetItems()
		allItems = append(allItems, items...)

		total := userList.GetPageInfo().GetTotal()
		from += int64(len(items))
		if len(items) == 0 || from >= total {
			break
		}
	}

	filter := makeFilter(&userLookup)
	var objects []attr.Value

	for _, item := range allItems {
		email := cast.ToString(item.GetEmail())
		firstName := cast.ToString(item.GetFirstName())
		lastName := cast.ToString(item.GetLastName())
		displayName := item.GetDisplayName()

		if !filter.accept(displayName, email) {
			continue
		}

		userStatus := ""
		if status := item.GetUserStatus(); status != nil {
			userStatus = string(*status)
		}
		userImportType := ""
		if importType := item.GetUserImportType(); importType != nil {
			userImportType = string(*importType)
		}

		obj, diags := types.ObjectValue(
			attrTypes,
			map[string]attr.Value{
				tfID:             types.StringValue(item.GetID()),
				tfEmail:          types.StringValue(email),
				tfFirstName:      types.StringValue(firstName),
				tfLastName:       types.StringValue(lastName),
				tfUserStatus:     types.StringValue(userStatus),
				tfUserImportType: types.StringValue(userImportType),
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

	userLookup.Items = list
	if diags := resp.State.Set(ctx, &userLookup); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

// makeFilter prepares lookup maps for display names and emails;
// if IsNull() then the respective map is nil
func makeFilter(userLookup *Lookup) *userFilter {
	var f userFilter

	if !userLookup.NameFilter.IsNull() {
		f.namesMap = make(map[string]struct{})
		for _, value := range userLookup.NameFilter.Elements() {
			valueStr := strings.Trim(value.String(), "\"") // Trim quotes
			f.namesMap[valueStr] = struct{}{}
		}
	}
	if !userLookup.EmailFilter.IsNull() {
		f.emailsMap = make(map[string]struct{})
		for _, value := range userLookup.EmailFilter.Elements() {
			valueStr := strings.Trim(value.String(), "\"") // Trim quotes
			f.emailsMap[valueStr] = struct{}{}
		}
	}
	return &f
}

// accept returns true if the item should be returned
func (f *userFilter) accept(name, email string) bool {
	if f.emailsMap == nil {
		if f.namesMap == nil {
			return true // no filters applied
		}
		if _, ok := f.namesMap[name]; ok {
			return true // name matches the filter
		}
		return false // name did not match
	}

	if _, ok := f.emailsMap[email]; ok {
		return true // email matches the filter
	}
	if f.namesMap == nil {
		return false // email did not match and names filter not used
	}
	if _, ok := f.namesMap[name]; ok {
		return true // name matches the filter
	}
	return false // nor email nor name matched
}

func importTypeFilter(value types.String) *cato_models.UserFilterInput {
	var apiFilter *cato_models.UserFilterInput
	if !value.IsNull() && value.ValueString() != "" {
		apiFilter = &cato_models.UserFilterInput{
			UserImportType: []*cato_models.UserImportTypeFilterInput{
				{
					In: []cato_models.UserImportType{
						cato_models.UserImportType(value.ValueString()),
					},
				},
			},
		}
	}

	return apiFilter
}
