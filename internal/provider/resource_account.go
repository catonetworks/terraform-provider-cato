package provider

import (
	"context"
	"encoding/json"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &accountResource{}
	_ resource.ResourceWithConfigure   = &accountResource{}
	_ resource.ResourceWithImportState = &accountResource{}
)

func NewAccountResource() resource.Resource {
	return &accountResource{}
}

type accountResource struct {
	client *catoClientData
}

func (r *accountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (r *accountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_account` resource contains the configuration parameters necessary to manage a Cato Networks account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Account ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Account name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Account description",
				Optional:    true,
			},
			"tenancy": schema.StringAttribute{
				Description: "Tenancy type (SINGLE_TENANT or MULTI_TENANT)",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("SINGLE_TENANT"),
				Validators: []validator.String{
					stringvalidator.OneOf("SINGLE_TENANT", "MULTI_TENANT"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone of the account (e.g., Africa/Accra)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Africa/Accra", "Africa/Addis_Ababa", "Africa/Algiers", "Africa/Asmara", "Africa/Bamako", "Africa/Bangui", "Africa/Banjul", "Africa/Bissau", "Africa/Bujumbura", "Africa/Cairo", "Africa/Casablanca", "Africa/Ceuta", "Africa/Conakry", "Africa/Dakar", "Africa/Dar_es_Salaam", "Africa/Djibouti", "Africa/Douala", "Africa/El_Aaiun", "Africa/Freetown", "Africa/Gaborone", "Africa/Harare", "Africa/Johannesburg", "Africa/Juba", "Africa/Kampala", "Africa/Khartoum", "Africa/Kigali", "Africa/Lagos", "Africa/Libreville", "Africa/Lome", "Africa/Luanda", "Africa/Lusaka", "Africa/Malabo", "Africa/Maputo", "Africa/Maseru", "Africa/Mbabane", "Africa/Mogadishu", "Africa/Monrovia", "Africa/Nairobi", "Africa/Ndjamena", "Africa/Niamey", "Africa/Nouakchott", "Africa/Ouagadougou", "Africa/Porto-Novo", "Africa/Tripoli", "Africa/Tunis", "Africa/Windhoek", "America/Anchorage", "America/Anguilla", "America/Antigua", "America/Argentina/Buenos_Aires", "America/Argentina/Catamarca", "America/Argentina/Cordoba", "America/Argentina/Jujuy", "America/Argentina/La_Rioja", "America/Argentina/Mendoza", "America/Argentina/Rio_Gallegos", "America/Argentina/Salta", "America/Argentina/San_Juan", "America/Argentina/San_Luis", "America/Argentina/Tucuman", "America/Argentina/Ushuaia", "America/Aruba", "America/Asuncion", "America/Atikokan", "America/Barbados", "America/Belize", "America/Blanc-Sablon", "America/Bogota", "America/Cambridge_Bay", "America/Caracas", "America/Cayenne", "America/Cayman", "America/Chicago", "America/Chihuahua", "America/Costa_Rica", "America/Creston", "America/Danmarkshavn", "America/Dawson", "America/Dawson_Creek", "America/Denver", "America/Dominica", "America/Edmonton", "America/El_Salvador", "America/Fort_Nelson", "America/Glace_Bay", "America/Godthab", "America/Goose_Bay", "America/Grand_Turk", "America/Grenada", "America/Guadeloupe", "America/Guatemala", "America/Guayaquil", "America/Guyana", "America/Halifax", "America/Havana", "America/Inuvik", "America/Iqaluit", "America/Jamaica", "America/La_Paz", "America/Lima", "America/Los_Angeles", "America/Lower_Princes", "America/Managua", "America/Martinique", "America/Miquelon", "America/Moncton", "America/Monterrey", "America/Montevideo", "America/Montserrat", "America/Nassau", "America/New_York", "America/Panama", "America/Pangnirtung", "America/Paramaribo", "America/Phoenix", "America/Port_of_Spain", "America/Port-au-Prince", "America/Puerto_Rico", "America/Punta_Arenas", "America/Rainy_River", "America/Rankin_Inlet", "America/Regina", "America/Resolute", "America/Santiago", "America/Santo_Domingo", "America/Scoresbysund", "America/St_Johns", "America/St_Kitts", "America/St_Lucia", "America/St_Thomas", "America/St_Vincent", "America/Swift_Current", "America/Tegucigalpa", "America/Thule", "America/Thunder_Bay", "America/Toronto", "America/Tortola", "America/Vancouver", "America/Whitehorse", "America/Winnipeg", "America/Yellowknife", "Arctic/Longyearbyen", "Asia/Aden", "Asia/Almaty", "Asia/Amman", "Asia/Anadyr", "Asia/Aqtau", "Asia/Aqtobe", "Asia/Ashgabat", "Asia/Atyrau", "Asia/Baghdad", "Asia/Bahrain", "Asia/Baku", "Asia/Bangkok", "Asia/Barnaul", "Asia/Beirut", "Asia/Bishkek", "Asia/Brunei", "Asia/Chita", "Asia/Choibalsan", "Asia/Colombo", "Asia/Damascus", "Asia/Dhaka", "Asia/Dubai", "Asia/Dushanbe", "Asia/Famagusta", "Asia/Ho_Chi_Minh", "Asia/Hong_Kong", "Asia/Hovd", "Asia/Hovd", "Asia/Irkutsk", "Asia/Jakarta", "Asia/Jayapura", "Asia/Jerusalem", "Asia/Kabul", "Asia/Kamchatka", "Asia/Karachi", "Asia/Kathmandu", "Asia/Khandyga", "Asia/Krasnoyarsk", "Asia/Kuala_Lumpur", "Asia/Kuching", "Asia/Kuwait", "Asia/Macau", "Asia/Magadan", "Asia/Makassar", "Asia/Manila", "Asia/Muscat", "Asia/Novokuznetsk", "Asia/Novosibirsk", "Asia/Omsk", "Asia/Oral", "Asia/Phnom_Penh", "Asia/Pontianak", "Asia/Pyongyang", "Asia/Qatar", "Asia/Qyzylorda", "Asia/Riyadh", "Asia/Sakhalin", "Asia/Samarkand", "Asia/Seoul", "Asia/Shanghai", "Asia/Singapore", "Asia/Srednekolymsk", "Asia/Taipei", "Asia/Tashkent", "Asia/Tbilisi", "Asia/Tehran", "Asia/Thimphu", "Asia/Tokyo", "Asia/Tomsk", "Asia/Ulaanbaatar", "Asia/Ulaanbaatar", "Asia/Urumqi", "Asia/Ust-Nera", "Asia/Vientiane", "Asia/Vladivostok", "Asia/Yakutsk", "Asia/Yangon", "Asia/Yekaterinburg", "Asia/Yerevan", "Atlantic/Azores", "Atlantic/Bermuda", "Atlantic/Canary", "Atlantic/Faroe", "Atlantic/Madeira", "Atlantic/Reykjavik", "Atlantic/South_Georgia", "Atlantic/Stanley", "Australia/ACT", "Australia/North", "Australia/NSW", "Australia/Queensland", "Australia/South", "Australia/Tasmania", "Australia/Victoria", "Australia/West", "Canada/Mountain", "Europe/Amsterdam", "Europe/Andorra", "Europe/Astrakhan", "Europe/Athens", "Europe/Belgrade", "Europe/Berlin", "Europe/Brussels", "Europe/Bucharest", "Europe/Budapest", "Europe/Busingen", "Europe/Chisinau", "Europe/Copenhagen", "Europe/Dublin", "Europe/Gibraltar", "Europe/Guernsey", "Europe/Helsinki", "Europe/Isle_of_Man", "Europe/Istanbul", "Europe/Jersey", "Europe/Kaliningrad", "Europe/Kiev", "Europe/Kirov", "Europe/Lisbon", "Europe/Ljubljana", "Europe/London", "Europe/Luxembourg", "Europe/Madrid", "Europe/Malta", "Europe/Minsk", "Europe/Monaco", "Europe/Moscow", "Europe/Nicosia", "Europe/Oslo", "Europe/Paris", "Europe/Podgorica", "Europe/Prague", "Europe/Riga", "Europe/Rome", "Europe/Samara", "Europe/San_Marino", "Europe/Sarajevo", "Europe/Saratov", "Europe/Skopje", "Europe/Sofia", "Europe/Stockholm", "Europe/Tallinn", "Europe/Tirane", "Europe/Ulyanovsk", "Europe/Uzhgorod", "Europe/Vaduz", "Europe/Vienna", "Europe/Vilnius", "Europe/Volgograd", "Europe/Warsaw", "Europe/Zagreb", "Europe/Zaporozhye", "Europe/Zurich", "Indian/Antananarivo", "Indian/Christmas", "Indian/Comoro", "Indian/Kerguelen", "Indian/Mahe", "Indian/Maldives", "Indian/Mauritius", "Indian/Mayotte", "Mexico/BajaNorte", "Pacific/Apia", "Pacific/Auckland", "Pacific/Bougainville", "Pacific/Chatham", "Pacific/Chuuk", "Pacific/Easter", "Pacific/Efate", "Pacific/Enderbury", "Pacific/Fakaofo", "Pacific/Fiji", "Pacific/Funafuti", "Pacific/Gambier", "Pacific/Guadalcanal", "Pacific/Guam", "Pacific/Honolulu", "Pacific/Kiritimati", "Pacific/Kosrae", "Pacific/Kwajalein", "Pacific/Majuro", "Pacific/Marquesas", "Pacific/Nauru", "Pacific/Niue", "Pacific/Norfolk", "Pacific/Noumea", "Pacific/Pago_Pago", "Pacific/Palau", "Pacific/Pitcairn", "Pacific/Pohnpei", "Pacific/Rarotonga", "Pacific/Saipan", "Pacific/Tahiti", "Pacific/Tarawa", "Pacific/Tongatapu", "Pacific/Wallis", "UTC-3", "UTC-4", "UTC-5", "UTC+05:30"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Account type (e.g., CUSTOMER)",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("CUSTOMER"),
				Validators: []validator.String{
					stringvalidator.OneOf("CUSTOMER", "PARTNER"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *accountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *accountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *accountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan Account
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.AddAccountInput{}

	if !plan.Description.IsNull() || !plan.Description.IsUnknown() {
		input.Description = plan.Description.ValueStringPointer()
	}

	if !plan.Name.IsNull() || !plan.Name.IsUnknown() {
		input.Name = plan.Name.ValueString()
	}

	if !plan.Tenancy.IsNull() || !plan.Tenancy.IsUnknown() {
		tenancyStr := plan.Tenancy.ValueString()
		input.Tenancy = cato_models.AccountTenancy(tenancyStr)
	}

	if !plan.Timezone.IsNull() || !plan.Timezone.IsUnknown() {
		input.Timezone = plan.Timezone.ValueString()
	}

	if !plan.Type.IsNull() || !plan.Type.IsUnknown() {
		input.Type = cato_models.AccountProfileType(plan.Type.ValueString())
	}

	//publishing new section
	tflog.Info(ctx, "Create.account")
	tflog.Debug(ctx, "Create.Account.request", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	addAccountResult, err := r.client.catov2.AccountManagementAddAccount(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.Account.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(addAccountResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			err.Error(),
		)
		return
	}

	// Validate account creation result
	if addAccountResult.AccountManagement.AddAccount == nil || addAccountResult.AccountManagement.AddAccount.ID == "" {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			"Account creation failed - no account data returned",
		)
		return
	}

	// Use the account data directly from the create response
	createdAccount := addAccountResult.AccountManagement.AddAccount
	tflog.Debug(ctx, "Account created successfully", map[string]interface{}{
		"account_id":   createdAccount.ID,
		"account_data": utils.InterfaceToJSONString(createdAccount),
	})

	// Update the state with the complete account information from the create response
	plan.Id = types.StringValue(createdAccount.ID)
	plan.Name = types.StringValue(createdAccount.Name)
	plan.Tenancy = types.StringValue(string(createdAccount.Tenancy))
	plan.Timezone = types.StringValue(createdAccount.TimeZone)
	plan.Type = types.StringValue(string(createdAccount.Type))

	if createdAccount.Description != nil {
		plan.Description = types.StringValue(*createdAccount.Description)
	} else {
		plan.Description = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state Account
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readAccountResponse, err := r.client.catov2.AccountManagement(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Read.Account.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(readAccountResponse),
	})

	if readAccountResponse.AccountManagement.Account.Description != nil {
		state.Description = types.StringValue(*readAccountResponse.AccountManagement.Account.Description)
	} else {
		state.Description = types.StringNull()
	}

	state.Name = types.StringValue(readAccountResponse.AccountManagement.Account.Name)
	state.Tenancy = types.StringValue(string(readAccountResponse.AccountManagement.Account.Tenancy))
	state.Timezone = types.StringValue(readAccountResponse.AccountManagement.Account.TimeZone)
	state.Id = types.StringValue(readAccountResponse.AccountManagement.Account.ID)
	state.Type = types.StringValue(string(readAccountResponse.AccountManagement.Account.Type))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan Account
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Account
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if any immutable fields have changed
	if !plan.Name.Equal(state.Name) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Account name cannot be changed after creation. Current: "+state.Name.ValueString()+", Requested: "+plan.Name.ValueString(),
		)
		return
	}

	if !plan.Tenancy.Equal(state.Tenancy) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Account tenancy cannot be changed after creation. Current: "+state.Tenancy.ValueString()+", Requested: "+plan.Tenancy.ValueString(),
		)
		return
	}

	if !plan.Timezone.Equal(state.Timezone) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Account timezone cannot be changed after creation. Current: "+state.Timezone.ValueString()+", Requested: "+plan.Timezone.ValueString(),
		)
		return
	}

	if !plan.Type.Equal(state.Type) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Account type cannot be changed after creation. Current: "+state.Type.ValueString()+", Requested: "+plan.Type.ValueString(),
		)
		return
	}

	// Check if description has actually changed
	if plan.Description.Equal(state.Description) {
		// No changes needed, just return the current state
		diags = resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		return
	}

	input := cato_models.UpdateAccountInput{}

	if !plan.Description.IsNull() || !plan.Description.IsUnknown() {
		input.Description = plan.Description.ValueStringPointer()
	}

	//publishing new section
	tflog.Info(ctx, "Update.account")
	tflog.Debug(ctx, "Update.Account.request", map[string]interface{}{
		"response": utils.InterfaceToJSONString(input),
	})
	updateAccountResult, err := r.client.catov2.AccountManagementUpdateAccount(ctx, input, plan.Id.ValueString())
	tflog.Debug(ctx, "Update.Account.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateAccountResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			err.Error(),
		)
		return
	}

	// Validate account update result
	if updateAccountResult.AccountManagement.UpdateAccount == nil {
		resp.Diagnostics.AddError(
			"Catov2 API Account error",
			"Account update failed - no account data returned",
		)
		return
	}

	// Use the account data directly from the update response
	updatedAccount := updateAccountResult.AccountManagement.UpdateAccount
	tflog.Debug(ctx, "Account updated successfully", map[string]interface{}{
		"account_id":   updatedAccount.ID,
		"account_data": utils.InterfaceToJSONString(updatedAccount),
	})

	// Update the state with the complete account information from the update response
	plan.Id = types.StringValue(updatedAccount.ID)
	plan.Name = types.StringValue(updatedAccount.Name)
	plan.Tenancy = types.StringValue(string(updatedAccount.Tenancy))
	plan.Timezone = types.StringValue(updatedAccount.TimeZone)
	plan.Type = types.StringValue(string(updatedAccount.Type))

	if updatedAccount.Description != nil {
		plan.Description = types.StringValue(*updatedAccount.Description)
	} else {
		plan.Description = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *accountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state Account
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Delete.AccountManagementRemoveAccount.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(state.Id),
	})
	removedAccountResult, err := r.client.catov2.AccountManagementRemoveAccount(ctx, state.Id.ValueString(), r.client.AccountId)
	tflog.Debug(ctx, "Delete.AccountManagementRemoveAccount.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removedAccountResult),
	})
	if err != nil {
		var apiError struct {
			NetworkErrors interface{} `json:"networkErrors"`
			GraphqlErrors []struct {
				Message string   `json:"message"`
				Path    []string `json:"path"`
			} `json:"graphqlErrors"`
		}
		isDisabled := false
		if parseErr := json.Unmarshal([]byte(err.Error()), &apiError); parseErr == nil && len(apiError.GraphqlErrors) > 0 {
			if apiError.GraphqlErrors[0].Message == "Please disable account before deleting" {
				isDisabled = true
			}
		}
		tflog.Debug(ctx, "Checking for account not ready for deletion, if not expired yet, gracefully failing deleting resource from state.")
		if isDisabled {
			resp.Diagnostics.AddError(
				"Catov2 API error",
				err.Error(),
			)
			return
		}
	}

}
