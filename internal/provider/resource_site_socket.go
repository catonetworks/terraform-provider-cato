package provider

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/dhcp"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

const (
	socketConnectionTypeAWS1500  = "SOCKET_AWS1500"
	socketConnectionTypeAZ1500   = "SOCKET_AZ1500"
	socketConnectionTypeESX1500  = "SOCKET_ESX1500"
	socketConnectionTypeX1600    = "SOCKET_X1600"
	socketConnectionTypeX1600LTE = "SOCKET_X1600_LTE"
	socketConnectionTypeX1700    = "SOCKET_X1700"
	socketConnectionTypeVGXAWS   = "VSOCKET_VGX_AWS"
	socketConnectionTypeVGXAzure = "VSOCKET_VGX_AZURE"
	socketConnectionTypeVGXESX   = "VSOCKET_VGX_ESX"
	socketInterfaceDestTypeLAN   = "LAN"
	socketCreateHydrationRetries = 6
)

var (
	_ resource.Resource                = &socketSiteResource{}
	_ resource.ResourceWithConfigure   = &socketSiteResource{}
	_ resource.ResourceWithImportState = &socketSiteResource{}
)

func NewSocketSiteResource() resource.Resource {
	return &socketSiteResource{}
}

func stringPointerForOptionalInput(value types.String) *string {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil
	}
	return value.ValueStringPointer()
}

type socketSiteResource struct {
	client           *catoClientData
	socketSiteClient SocketSiteClient
}

type SocketSiteClient interface {
	SiteAddSocketSite(ctx context.Context, addSocketSiteInput cato_models.AddSocketSiteInput, accountID string,
		interceptors ...clientv2.RequestInterceptor) (*cato_go_sdk.SiteAddSocketSite, error)
}

type nativeInterfaceDetails struct {
	interfaceIndex *string
	interfaceID    *string
	interfaceName  *string
	destType       *string
}

var numberRE = regexp.MustCompile(`^\d+$`)

func (r *socketSiteResource) getSocketSiteClient() SocketSiteClient {
	if r.socketSiteClient != nil {
		return r.socketSiteClient
	}

	if r.client == nil {
		return nil
	}

	return r.client.catov2
}

func (r *socketSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_site"
}

func (r *socketSiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_socket_site` resource contains the configuration parameters necessary to add a socket site " +
			"to the Cato cloud ([virtual socket in AWS/Azure, or physical socket]" +
			"(https://support.catonetworks.com/hc/en-us/articles/4413280502929-Working-with-X1500-X1600-and-X1700-Socket-Sites)). " +
			"Documentation for the underlying API used in this resource can be found at " +
			"[mutation.addSocketSite()](https://api.catonetworks.com/documentation/#mutation-site.addSocketSite). \n\n" +
			"**Note**: For AWS deployments, please accept the [EULA for the Cato Networks AWS Marketplace product]" +
			"(https://aws.amazon.com/marketplace/pp?sku=dvfhly9fuuu67tw59c7lt5t3c).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Site ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Site name",
				Required:    true,
			},
			"connection_type": schema.StringAttribute{
				Description:   "Connection type for the site (SOCKET_X1500, SOCKET_AWS1500, SOCKET_AZ1500, ...)",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{validators.SiteConnectionTypeValidator{}},
			},
			"site_type": schema.StringAttribute{
				Description: "Site type (https://api.catonetworks.com/documentation/#definition-SiteType)",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Site description",
				Optional:    true,
			},
			"native_range":  r.schemaNativeRange(),
			"site_location": r.schemaSiteLocation(),
			"sockets":       r.schemaSockets(),
		},
	}
}

func (r *socketSiteResource) schemaNativeRange() schema.SingleNestedAttribute { //nolint:funlen
	return schema.SingleNestedAttribute{
		Description: "Site lan native range settings",
		Required:    true,
		Validators:  []validator.Object{validators.GetNativeRangeValidator()},
		Attributes: map[string]schema.Attribute{
			"interface_index": schema.StringAttribute{
				Description: "LAN native range interface index, default is LAN1 for SOCKET_X1500 models, " +
					"INT_5 for SOCKET_X1600 and SOCKET_X1600_LTE, and INT_3 for SOCKET_X1700 models",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{validators.SocketInterfaceIndexValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"interface_id": schema.StringAttribute{
				Description:   "LAN native range interface id",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"native_network_range": schema.StringAttribute{
				Description: "Site native IP range (CIDR)",
				Required:    true,
			},
			"native_network_lan_interface_id": schema.StringAttribute{
				Description:   "ID of native range LAN interface (for additional network range update purposes)",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"native_network_range_id": schema.StringAttribute{
				Description:   "Site native IP range ID (for update purpose)",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"interface_name": schema.StringAttribute{
				Description:   "LAN native range interface name (e.g., 'LAN 01')",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"range_name": schema.StringAttribute{
				Description:   "Native range name (typically 'Native Range')",
				Computed:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"range_id": schema.StringAttribute{
				Description:   "Native range ID (base64 encoded identifier)",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"gateway": schema.StringAttribute{
				Description:   "Gateway IP address for the native range",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"vlan": schema.Int64Attribute{
				Description: "VLAN ID for the site native range (optional)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"mdns_reflector": schema.BoolAttribute{
				Description: "Site native range mDNS reflector. When enabled, the Socket functions as an mDNS gateway, " +
					"it relays mDNS requests and response between all enabled subnets.",
				Optional:      true,
				Computed:      true,
				Default:       booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"local_ip": schema.StringAttribute{
				Description: "Site native range local ip",
				Required:    true,
			},
			"primary_management_ip": schema.StringAttribute{
				Description:   "Site native range primary management IP",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"secondary_management_ip": schema.StringAttribute{
				Description:   "Site native range secondary management IP",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"range_type": schema.StringAttribute{
				Description:   "NATIVE",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"translated_subnet": schema.StringAttribute{
				Description:   "Site translated native IP range (CIDR)",
				Computed:      true,
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"lag_min_links": schema.Int64Attribute{
				Description: "Number of interfaces to include in the link aggregation, " +
					"only relevant for LAN_LAG_MASTER and LAN_LAG_MASTER_AND_VRRP interface destination types",
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(1)},
			},
			"interface_dest_type": schema.StringAttribute{
				Description: "Socket interface destination type for the native interface, " +
					"example values: LAN, LAN_LAG_MASTER, LAN_LAG_MASTER_AND_VRRP, LAN_AND_HA, VRRP, VRRP_AND_LAN",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{validators.SocketInterfaceDestTypeValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dhcp_settings": dhcp.SchemaDhcpSettings(false),
		},
	}
}

func (r *socketSiteResource) schemaSiteLocation() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Site location",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"country_code": schema.StringAttribute{
				Description: "Site country code (can be retrieve from entityLookup)",
				Required:    true,
			},
			"state_code": schema.StringAttribute{
				Description: "Optionnal site state code(can be retrieve from entityLookup)",
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Site timezone (can be retrieve from entityLookup)",
				Required:    true,
			},
			"city": schema.StringAttribute{
				Description:   "Optionnal city",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"address": schema.StringAttribute{
				Description:   "Optionnal address",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *socketSiteResource) schemaSockets() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description:   "Socket information",
		Computed:      true,
		PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Description:   "Socket ID",
					Computed:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
				"serial_number": schema.StringAttribute{
					Description:   "Socket serial number",
					Computed:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
				"is_primary": schema.BoolAttribute{
					Description:   "Indicates if the socket is primary",
					Computed:      true,
					PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
				},
				"platform": schema.StringAttribute{
					Description:   "Socket platform",
					Computed:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
			},
		},
	}
}

func (r *socketSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *socketSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// TODO: check if this is really necessary
	// Call Read to hydrate the full state from the API
	readReq := resource.ReadRequest{State: resp.State}
	readResp := resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, &readResp)

	// Copy diagnostics and state back to the import response
	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State
}

// Create cato_socket_site resource
//
//nolint:funlen // create flow composes multiple API calls and hydration checks in sequence.
func (r *socketSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, cfg tf.SocketSite
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a socket site - API call
	siteID := r.createBasicSocketSite(ctx, &plan, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Find Network Range ID for the created site and update the NetworkRange details
	networkRange := r.findNativeRange(ctx, siteID, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	r.updateNetworkRange(ctx, &plan, networkRange.GetNetworkRangeID(), false, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Move interface to a non-default index if specified in the plan
	defaultInterfaceIndex := tf.InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]
	r.assignInterfaceIndex(ctx, defaultInterfaceIndex, &plan, siteID, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// Update Socket Interface
	r.updateSocketInterface(ctx, &plan, siteID, false, &diags)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// hydrate the state with API data, retrying briefly for eventual consistency
	var (
		hydratedState tf.SocketSite
		siteExists    bool
	)
	for attempt := 0; attempt < socketCreateHydrationRetries; attempt++ {
		attemptDiags := diag.Diagnostics{}
		hydratedState, siteExists = r.hydrateSocketSiteState(ctx, &cfg, plan, siteID, &attemptDiags)
		if attemptDiags.HasError() {
			if attempt == socketCreateHydrationRetries-1 {
				resp.Diagnostics.Append(attemptDiags...)
				return
			}
			time.Sleep(2 * time.Second)
			continue
		}
		if siteExists {
			break
		}
		time.Sleep(2 * time.Second)
	}

	// if the site is still not visible, keep minimal state to avoid losing created resource
	if !siteExists {
		fallbackState := cfg
		fallbackState.ID = types.StringValue(siteID)
		tflog.Warn(ctx, "site not found after create; preserving minimal state until next refresh")
		diags = resp.State.Set(ctx, &fallbackState)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overiding state with socket site id
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), siteID)
}

// Read cato_socket_site resource
func (r *socketSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tf.SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists := r.hydrateSocketSiteState(ctx, nil, state, state.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update cato_socket_site resource
func (r *socketSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var cfg, plan, state tf.SocketSite
	var planNativeRange, stateNativeRange tf.NativeRange

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	diags = req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := plan.ID.ValueString()
	isHA := r.isHA(&state)

	// Update general socket site details
	r.updateBasicSocketSite(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update network range details
	if utils.CheckErr(&diags, plan.NativeRange.As(ctx, &planNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	networkRangeID := planNativeRange.NativeNetworkRangeID.ValueString()
	r.updateNetworkRange(ctx, &plan, networkRangeID, isHA, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Move interface to another index if needed
	if utils.CheckErr(&diags, state.NativeRange.As(ctx, &stateNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	currentInterfaceIndex := cato_models.SocketInterfaceIDEnum(stateNativeRange.InterfaceIndex.ValueString())
	r.assignInterfaceIndex(ctx, currentInterfaceIndex, &plan, siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Socket Interface
	r.updateSocketInterface(ctx, &plan, siteID, isHA, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// hydrate the state with API data
	hydratedState, siteExists := r.hydrateSocketSiteState(ctx, &cfg, plan, siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete cato_socket_site resource
func (r *socketSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tf.SocketSite
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId, cato_models.EntityType("site"),
		nil, nil, nil, nil, []string{state.ID.ValueString()}, nil, nil, nil)
	tflog.Debug(ctx, "Create.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// check if site exist before removing
	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
		_, err := r.client.catov2.SiteRemoveSite(ctx, state.ID.ValueString(), r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteRemoveSite error",
				err.Error(),
			)
			return
		}
	}
}

// hydrateSocketSiteState populates the tf.SocketSite state with data from API responses
func (r *socketSiteResource) hydrateSocketSiteState(ctx context.Context, cfg *tf.SocketSite, state tf.SocketSite,
	siteID string, diags *diag.Diagnostics,
) (newState tf.SocketSite, siteExists bool) {
	// Fetch account snapshot for this site
	siteAccountSnapshotAPIData, err := r.client.catov2.AccountSnapshot(ctx, []string{siteID}, nil, &r.client.AccountId)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to fetch AccountSnapshot for site '%s'", siteID), err.Error())
		return state, false
	}
	var siteSnapshot *cato_go_sdk.AccountSnapshot_AccountSnapshot_Sites
	for _, site := range siteAccountSnapshotAPIData.GetAccountSnapshot().GetSites() {
		if site.ID != nil && *site.ID == siteID {
			siteSnapshot = site
			break
		}
	}
	if siteSnapshot == nil || siteSnapshot.InfoSiteSnapshot == nil {
		diags.AddError(fmt.Sprintf("site with ID '%s' not found in account snapshot", siteID), "")
		return state, false
	}
	siteInfo := siteSnapshot.InfoSiteSnapshot

	// Fetch site general details for location data
	siteDetails, err := r.client.catov2.SiteGeneralDetails(ctx,
		cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID}, r.client.AccountId)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to fetch SiteGeneralDetails for site '%s'", siteID), err.Error())
		return state, true
	}

	// Fetch site native range and interface details
	networkRange := r.findNativeRange(ctx, siteID, diags)
	if diags.HasError() {
		return state, true
	}

	// Fetch network interface details
	defaultInterface := r.findDefaultInterface(ctx, siteID, state.NativeRange, diags)
	if diags.HasError() {
		return state, true
	}

	newState = tf.SocketSite{
		ID:             types.StringValue(siteID),
		Name:           types.StringPointerValue(siteInfo.GetName()),
		ConnectionType: r.fixConnectionType(siteInfo.GetConnType()),
		SiteType:       types.StringPointerValue((*string)(siteInfo.GetType())),
		Description:    types.StringPointerValue(siteInfo.GetDescription()),
		NativeRange:    r.parseNativeRange(ctx, cfg, networkRange, defaultInterface, state.NativeRange, diags),
		SiteLocation:   r.parseSiteLocation(ctx, siteDetails, state.SiteLocation, diags),
		Sockets:        r.parseSockets(ctx, siteInfo, diags),
	}
	if diags.HasError() {
		return state, true
	}

	return newState, true
}

// fixConnectionType translates any VSOCKET_VGX_* connection types to their SOCKET_* equivalents
// for backward compatibility with existing state/plan values
func (r *socketSiteResource) fixConnectionType(connTypeFromAPI *cato_models.ProtoType) types.String {
	if connTypeFromAPI == nil || *connTypeFromAPI == "" {
		return types.StringNull()
	}
	switch *connTypeFromAPI {
	case cato_models.ProtoTypeVsocketVgxAWS:
		return types.StringValue(string(cato_models.SiteConnectionTypeEnumSocketAWS1500))
	case cato_models.ProtoTypeVsocketVgxAzure:
		return types.StringValue(string(cato_models.SiteConnectionTypeEnumSocketAz1500))
	case cato_models.ProtoTypeVsocketVgxEsx:
		return types.StringValue(string(cato_models.SiteConnectionTypeEnumSocketEsx1500))
	}
	return types.StringValue(string(*connTypeFromAPI))
}

// parseSiteLocation converts API site location data to the types.Object or tf.SiteLocation
func (r *socketSiteResource) parseSiteLocation(ctx context.Context, siteGenDetails *cato_go_sdk.SiteGeneralDetails,
	siteLocation types.Object, diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	var planLocation tf.SiteLocation

	if siteGenDetails == nil || siteGenDetails.Site.SiteGeneralDetails == nil {
		return types.ObjectNull(tf.SiteLocationResourceAttrTypes)
	}
	siteLoc := siteGenDetails.Site.SiteGeneralDetails.SiteLocation

	// Prepare site location object
	state := siteLoc.GetStateCode()
	if state != nil && *state == "" {
		state = nil
	}
	tfLocation := tf.SiteLocation{
		CountryCode: types.StringValue(siteLoc.GetCountryCode()),
		StateCode:   types.StringPointerValue(state),
		Timezone:    types.StringValue(siteLoc.GetTimezone()),
		Address:     types.StringPointerValue(siteLoc.GetAddress()),
		City:        types.StringPointerValue(siteLoc.GetCityName()),
	}

	// API sometimes returns empty string and sometimes null for city - use what is in the plan in that case
	city := siteLoc.GetCityName()
	if ((city == nil) || (*city == "")) && utils.HasValue(siteLocation) {
		if utils.CheckErr(diags, siteLocation.As(ctx, &planLocation, basetypes.ObjectAsOptions{})) {
			return types.ObjectNull(tf.SiteLocationResourceAttrTypes)
		}
		if planLocation.City.IsNull() || (utils.HasValue(planLocation.City) && (planLocation.City.ValueString() == "")) {
			tfLocation.City = planLocation.City
		}
	}

	locObj, objDiags := types.ObjectValueFrom(ctx, tf.SiteLocationResourceAttrTypes, tfLocation)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PreferredPopLocationModelTypes)
	}

	return locObj
}

// parseSockets converts API site socket data to the types.Set of tf.Socket
func (r *socketSiteResource) parseSockets(ctx context.Context, siteInfo *cato_go_sdk.AccountSnapshot_AccountSnapshot_Sites_InfoSiteSnapshot,
	diags *diag.Diagnostics,
) types.Set {
	var objDiags diag.Diagnostics
	setNull := types.SetNull(types.ObjectType{AttrTypes: tf.SocketTypes})

	if siteInfo == nil || len(siteInfo.Sockets) == 0 {
		return setNull
	}

	socketObjects := make([]types.Object, 0, len(siteInfo.Sockets))

	for _, soc := range siteInfo.Sockets {
		if soc == nil {
			continue
		}

		// ProtocolPorts item
		tfSocket := tf.Socket{
			ID:           types.StringPointerValue(soc.ID),
			SerialNumber: types.StringPointerValue(soc.Serial),
			IsPrimary:    types.BoolPointerValue(soc.IsPrimary),
			Platform:     types.StringPointerValue((*string)(soc.PlatformSocketInfo)),
		}
		tfSocketObj, objDiags := types.ObjectValueFrom(ctx, tf.SocketTypes, tfSocket)
		diags.Append(objDiags...)
		if diags.HasError() {
			return setNull
		}
		socketObjects = append(socketObjects, tfSocketObj)
	}

	// convert slice to types.Set
	tfSocketSet, objDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: tf.SocketTypes}, socketObjects)
	diags.Append(objDiags...)
	if diags.HasError() {
		return setNull
	}

	return tfSocketSet
}

// checkDhcpSettingsDefault checks DHCP settings config or state,
// If config is provided,
// return true if it is not defined or if dhcp_type is set to ACCOUNT_DEFAULT, false otherwise
// If config is not provided (nil), check the state with the same logic to determine if DHCP settings are default
//
// Reason: API does not return ACCOUNT_DEFAULT, but some other value based on CMA account configuration
func (r *socketSiteResource) checkDhcpSettingsDefault(ctx context.Context, cfg *tf.SocketSite,
	stateNativeRange types.Object, diags *diag.Diagnostics,
) (isDhcpSettingsDefault bool) {
	if cfg != nil {
		var cfgNativeRange tf.NativeRange
		if utils.CheckErr(diags, cfg.NativeRange.As(ctx, &cfgNativeRange, basetypes.ObjectAsOptions{})) {
			return false
		}
		if cfgNativeRange.DhcpSettings.IsNull() {
			return true
		}

		var cfgDhcpSettings tf.DhcpSettings
		if utils.CheckErr(diags, cfgNativeRange.DhcpSettings.As(ctx, &cfgDhcpSettings, basetypes.ObjectAsOptions{})) {
			return false
		}
		if utils.HasValue(cfgDhcpSettings.DhcpType) &&
			(cfgDhcpSettings.DhcpType.ValueString() == string(cato_models.DhcpTypeAccountDefault)) {
			return true
		}
		return false
	}

	// cfg is nil -> called from Read(); check the state
	if !utils.HasValue(stateNativeRange) {
		return false
	}
	var tfStateNativeRange tf.NativeRange
	if utils.CheckErr(diags, stateNativeRange.As(ctx, &tfStateNativeRange, basetypes.ObjectAsOptions{})) {
		return false
	}
	if tfStateNativeRange.DhcpSettings.IsNull() {
		return true
	}
	var stateDhcpSettings tf.DhcpSettings
	if utils.CheckErr(diags, tfStateNativeRange.DhcpSettings.As(ctx, &stateDhcpSettings, basetypes.ObjectAsOptions{})) {
		return false
	}
	if utils.HasValue(stateDhcpSettings.DhcpType) &&
		(stateDhcpSettings.DhcpType.ValueString() == string(cato_models.DhcpTypeAccountDefault)) {
		return true
	}
	return false
}

// parseNativeRange converts API native range data to the types.Object or tf.NativeRange
func (r *socketSiteResource) parseNativeRange(ctx context.Context, cfg *tf.SocketSite,
	networkRange *cato_go_sdk.NetworkRangeList_Site_NetworkRangeList_Items,
	nativeInterface *nativeInterfaceDetails, stateNativeRange types.Object, diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	var tfStateNativeRange tf.NativeRange

	if networkRange == nil {
		return types.ObjectNull(tf.SiteNativeRangeResourceAttrTypes)
	}

	// DHCP settings
	isDhcpSettingsDefault := r.checkDhcpSettingsDefault(ctx, cfg, stateNativeRange, diags)
	dhcpSettingsObj := dhcp.SettingsDefault(ctx, diags)
	if networkRange.DhcpSettings != nil && !isDhcpSettingsDefault {
		dhcpSettingsObj = dhcp.ParseSettings(ctx, r.client,
			&cato_go_sdk.NetworkRange_Site_NetworkRange_DhcpSettings{
				DhcpMicrosegmentation: networkRange.DhcpSettings.DhcpMicrosegmentation,
				DhcpType:              networkRange.DhcpSettings.DhcpType,
				IPRange:               networkRange.DhcpSettings.IPRange,
				RelayGroupID:          networkRange.DhcpSettings.RelayGroupID,
			},
			diags)
	}
	if diags.HasError() {
		return types.ObjectNull(tf.SiteNativeRangeResourceAttrTypes)
	}

	if nativeInterface == nil {
		nativeInterface = &nativeInterfaceDetails{}
	}

	// decode status NativeRange
	if utils.HasValue(stateNativeRange) {
		if utils.CheckErr(diags, stateNativeRange.As(ctx, &tfStateNativeRange, basetypes.ObjectAsOptions{})) {
			return types.ObjectNull(tf.SiteNativeRangeResourceAttrTypes)
		}
	}

	// try to get LagMinLinks from state; not available in API; TODO: add to API
	lagMinLinks := types.Int64Null()
	if utils.HasValue(stateNativeRange) {
		lagMinLinks = tfStateNativeRange.LagMinLinks
	}

	// Fix interface index - API sometimes returns 'INT_5', sometimes just 5
	fixedIfaceIndex := nativeInterface.interfaceIndex
	if fixedIfaceIndex != nil && numberRE.MatchString(*fixedIfaceIndex) {
		fixedIfaceIndex = ptr("INT_" + *fixedIfaceIndex)
	}

	localIP := types.StringPointerValue(networkRange.LocalIP)
	if networkRange.LocalIP == nil { // In HA scenario the API does not return local IP
		if networkRange.PrimaryManagementIP != nil { // AWS: use primary management IP as local IP
			localIP = types.StringPointerValue(networkRange.PrimaryManagementIP)
		} else if utils.HasValue(stateNativeRange) { // try state local IP
			localIP = tfStateNativeRange.LocalIP
		}
	}

	// Prepare native range object
	tfNativeRange := tf.NativeRange{
		InterfaceIndex:              types.StringPointerValue(fixedIfaceIndex),
		InterfaceID:                 types.StringPointerValue(nativeInterface.interfaceID),
		InterfaceName:               types.StringPointerValue(nativeInterface.interfaceName),
		NativeNetworkLanInterfaceID: types.StringPointerValue(nativeInterface.interfaceID),

		NativeNetworkRange:    types.StringValue(networkRange.Subnet),
		NativeNetworkRangeID:  types.StringValue(networkRange.NetworkRangeID),
		RangeName:             types.StringValue(networkRange.Name),
		RangeID:               types.StringValue(networkRange.NetworkRangeID),
		LocalIP:               localIP,
		PrimaryManagementIP:   types.StringPointerValue(networkRange.PrimaryManagementIP),
		SecondaryManagementIP: types.StringPointerValue(networkRange.SecondaryManagementIP),
		TranslatedSubnet:      types.StringPointerValue(networkRange.TranslatedSubnet),
		Gateway:               types.StringPointerValue(networkRange.Gateway),
		RangeType:             types.StringValue(strings.ToUpper(string(networkRange.RangeType))),
		DhcpSettings:          dhcpSettingsObj,
		Vlan:                  types.Int64PointerValue(networkRange.Vlan),
		MdnsReflector:         types.BoolValue(networkRange.MdnsReflector),
		LagMinLinks:           lagMinLinks,
		InterfaceDestType:     types.StringPointerValue(nativeInterface.destType),
	}

	netRangeObj, objDiags := types.ObjectValueFrom(ctx, tf.SiteNativeRangeResourceAttrTypes, tfNativeRange)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(tf.SiteNativeRangeResourceAttrTypes)
	}

	return netRangeObj
}

// prepareSiteLocation constructs the API input (location part) for SiteAddSocketSite() from the Terraform plan data.
func (r *socketSiteResource) prepareSiteLocation(ctx context.Context, location types.Object, diags *diag.Diagnostics,
) *cato_models.AddSiteLocationInput {
	if !utils.HasValue(location) {
		return nil
	}

	var tfLocation tf.SiteLocation
	if utils.CheckErr(diags, location.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.AddSiteLocationInput{
		Address:     parse.KnownStringPointer(tfLocation.Address),
		City:        parse.KnownStringPointer(tfLocation.City),
		CountryCode: tfLocation.CountryCode.ValueString(),
		StateCode:   parse.KnownStringPointer(tfLocation.StateCode),
		Timezone:    tfLocation.Timezone.ValueString(),
	}
}

// prepareSocketSiteInput constructs the API input for SiteAddSocketSite() from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.
func (r *socketSiteResource) prepareSocketSiteInput(ctx context.Context, plan *tf.SocketSite, diags *diag.Diagnostics,
) *cato_models.AddSocketSiteInput {
	var tfNativeRange tf.NativeRange
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return nil
	}

	input := &cato_models.AddSocketSiteInput{
		ConnectionType:     cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString()),
		Description:        parse.KnownStringPointer(plan.Description),
		Name:               plan.Name.ValueString(),
		NativeNetworkRange: tfNativeRange.NativeNetworkRange.ValueString(),
		SiteLocation:       r.prepareSiteLocation(ctx, plan.SiteLocation, diags),
		SiteType:           cato_models.SiteType(plan.SiteType.ValueString()),
		TranslatedSubnet:   stringPointerForOptionalInput(tfNativeRange.TranslatedSubnet),
		Vlan:               (*scalars.Vlan)(parse.KnownInt64Pointer(tfNativeRange.Vlan)),
	}
	return input
}

// prepareNetworkRangeInput constructs the API input for SiteUpdateNetworkRange() from the Terraform plan data.
// It may add error(s) to the diagnostics if the input data is invalid.
func (r *socketSiteResource) prepareNetworkRangeInput(ctx context.Context, plan *tf.SocketSite, isHA bool, diags *diag.Diagnostics,
) *cato_models.UpdateNetworkRangeInput {
	var tfNativeRange tf.NativeRange
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return nil
	}
	input := &cato_models.UpdateNetworkRangeInput{
		Subnet:           parse.KnownStringPointer(tfNativeRange.NativeNetworkRange),
		TranslatedSubnet: stringPointerForOptionalInput(tfNativeRange.TranslatedSubnet),
		MdnsReflector:    parse.KnownBoolPointer(tfNativeRange.MdnsReflector),
		Vlan:             parse.KnownInt64Pointer(tfNativeRange.Vlan),
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, r.client, cato_models.SubnetTypeNative, tfNativeRange.DhcpSettings, diags),
		//    AzureFloatingIP *string `json:"azureFloatingIp,omitempty"` TODO: implement when AZURE HA support is added
	}

	if !isHA { // for HA scenario, local IP is not allowed to be modified
		input.LocalIP = parse.KnownStringPointer(tfNativeRange.LocalIP)
	}

	return input
}

// prepareSocketInterfaceInput prepares inputs for SiteUpdateSocketInterface API,
// on error updates diags
func (r *socketSiteResource) prepareSocketInterfaceInput(ctx context.Context, plan *tf.SocketSite, isHA bool, diags *diag.Diagnostics,
) (*cato_models.UpdateSocketInterfaceInput, cato_models.SocketInterfaceIDEnum) {
	var (
		tfNativeRange     tf.NativeRange
		lagLinksDestTypes = []cato_models.SocketInterfaceDestType{ // set lag input for these dest types
			cato_models.SocketInterfaceDestTypeLanLagMaster,
			cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp,
		}
		lanDestTypes = []cato_models.SocketInterfaceDestType{ // set lan input for these dest types
			cato_models.SocketInterfaceDestTypeLan,
			cato_models.SocketInterfaceDestTypeLanAndHa,
			cato_models.SocketInterfaceDestTypeVrrpAndLan,
			cato_models.SocketInterfaceDestTypeLanLagMaster,
			cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp,
		}
	)

	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return nil, ""
	}

	interfaceDestType := cato_models.SocketInterfaceDestType(tfNativeRange.InterfaceDestType.ValueString())

	if interfaceDestType == "" {
		interfaceDestType = cato_models.SocketInterfaceDestTypeLan // default to LAN if not specified
	}

	// Determine interface ifaceName with the following precedence: InterfaceName > InterfaceIndex > default based on connection type
	ifaceName := parse.KnownStringPointer(tfNativeRange.InterfaceName)
	if ifaceName == nil {
		ifaceName = parse.KnownStringPointer(tfNativeRange.InterfaceIndex)
	}
	if ifaceName == nil {
		ifaceName = ptr(string(tf.InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]))
	}

	input := &cato_models.UpdateSocketInterfaceInput{
		DestType: interfaceDestType,
		Name:     ifaceName,
	}

	// MinLinks for LAG
	if utils.HasValue(tfNativeRange.LagMinLinks) && slices.Contains(lagLinksDestTypes, interfaceDestType) {
		input.Lag = &cato_models.SocketInterfaceLagInput{MinLinks: tfNativeRange.LagMinLinks.ValueInt64()}
	}

	// LAN input
	if slices.Contains(lanDestTypes, interfaceDestType) && !isHA {
		input.Lan = &cato_models.SocketInterfaceLanInput{
			LocalIP:          tfNativeRange.LocalIP.ValueString(),
			Subnet:           tfNativeRange.NativeNetworkRange.ValueString(),
			TranslatedSubnet: stringPointerForOptionalInput(tfNativeRange.TranslatedSubnet),
		}
	}

	// get interface index if configured, otherwise use a default based on connection type
	interfaceIndex := cato_models.SocketInterfaceIDEnum(tfNativeRange.InterfaceIndex.ValueString())
	if interfaceIndex == "" {
		interfaceIndex = tf.InterfaceByConnType[cato_models.SiteConnectionTypeEnum(plan.ConnectionType.ValueString())]
	}

	return input, interfaceIndex
}

// createBasicSocketSite creates a socket site and returns the new site ID.
// It does NOT update native range or interface details.
func (r *socketSiteResource) createBasicSocketSite(ctx context.Context, plan *tf.SocketSite, diags *diag.Diagnostics) (siteID string) {
	socketSiteInput := r.prepareSocketSiteInput(ctx, plan, diags)
	if diags.HasError() {
		return ""
	}
	socketSite, err := r.getSocketSiteClient().SiteAddSocketSite(ctx, *socketSiteInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteAddSocketSite error", err.Error())
		return ""
	}
	siteID = socketSite.Site.AddSocketSite.GetSiteID()
	if siteID == "" {
		diags.AddError("Catov2 API SiteAddSocketSite error", "empty site ID returned from API")
		return ""
	}
	return siteID
}

// updateBasicSocketSite updates a socket site
// It does NOT update native range or interface details.
func (r *socketSiteResource) updateBasicSocketSite(ctx context.Context, plan *tf.SocketSite, diags *diag.Diagnostics) {
	input := cato_models.UpdateSiteGeneralDetailsInput{
		Description: parse.KnownStringPointer(plan.Description),
		Name:        plan.Name.ValueStringPointer(),
		SiteType:    (*cato_models.SiteType)(plan.SiteType.ValueStringPointer()),
	}
	if utils.HasValue(plan.SiteLocation) {
		var tfLocation tf.SiteLocation
		if utils.CheckErr(diags, plan.SiteLocation.As(ctx, &tfLocation, basetypes.ObjectAsOptions{})) {
			return
		}
		siteLocationInput := &cato_models.UpdateSiteLocationInput{
			Address:     parse.KnownStringPointer(tfLocation.Address),
			CityName:    parse.KnownStringPointer(tfLocation.City),
			CountryCode: tfLocation.CountryCode.ValueStringPointer(),
			StateCode:   parse.KnownStringPointer(tfLocation.StateCode),
			Timezone:    tfLocation.Timezone.ValueStringPointer(),
		}
		input.SiteLocation = siteLocationInput
	}

	_, err := r.client.catov2.SiteUpdateSiteGeneralDetails(ctx, plan.ID.ValueString(), input, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateSiteGeneralDetails error", err.Error())
		return
	}
}

// findNativeRange looks up the native network range for the given site ID and returns its details.
func (r *socketSiteResource) findNativeRange(ctx context.Context, siteID string, diags *diag.Diagnostics,
) (networkRange *cato_go_sdk.NetworkRangeList_Site_NetworkRangeList_Items) {
	input := cato_models.NetworkRangeListInput{Site: &cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID}}
	queryNetworkRangeResult, err := r.client.catov2.NetworkRangeList(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Catov2 API NetworkRangeList error", err.Error())
		return
	}
	for _, netRange := range queryNetworkRangeResult.GetSite().GetNetworkRangeList().GetItems() {
		if netRange.RangeType == cato_models.SubnetTypeNative {
			return netRange
		}
	}
	diags.AddError("Native network range not found", fmt.Sprintf("no native network range found for site ID %s", siteID))
	return nil
}

// findDefaultInterface looks up the default network interface for the given site ID and returns its details.
func (r *socketSiteResource) findDefaultInterface(ctx context.Context, siteID string,
	nativeRange types.Object, diags *diag.Diagnostics,
) *nativeInterfaceDetails {
	var defaultIface *cato_go_sdk.EntityLookup_EntityLookup_Items
	helperField := func(fieldName string, it *cato_go_sdk.EntityLookup_EntityLookup_Items) *string {
		valAny, ok := it.HelperFields[fieldName]
		if !ok {
			return nil
		}
		if valStr, ok := valAny.(string); ok {
			return &valStr
		}
		return nil
	}

	// Retrieve default interface range attributes
	siteEntity := &cato_models.EntityInput{Type: cato_models.EntityTypeSite, ID: siteID}
	result, err := r.client.catov2.EntityLookup(ctx, r.client.AccountId,
		cato_models.EntityTypeNetworkInterface, ptr(int64(0)), nil, siteEntity, nil, nil, nil, nil, nil)
	if err != nil {
		diags.AddError("Catov2 API EntityLookup 'networkInterface' error", err.Error())
		return nil
	}

	// find the default interface based on helper field "isDefault" set by API
	for _, item := range result.EntityLookup.GetItems() {
		if isDefaultAny, ok := item.HelperFields["isDefault"]; ok {
			if isDefault, ok := isDefaultAny.(bool); ok && isDefault {
				defaultIface = item
				break
			}
		}
	}

	// Falback: if API does not return "isDefault" as a helper field, try to match the interface ID from state
	if defaultIface == nil && utils.HasValue(nativeRange) {
		var tfNativeRange tf.NativeRange
		if utils.CheckErr(diags, nativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
			return nil
		}
		statIfaceID := tfNativeRange.InterfaceID.ValueString()
		for _, item := range result.EntityLookup.GetItems() {
			if item.Entity.ID == statIfaceID {
				defaultIface = item
				break
			}
		}
	}

	if defaultIface == nil {
		diags.AddError("Default interface not found", fmt.Sprintf("no default network interface found for site ID %s", siteID))
		return nil
	}

	ifaceDetails := &nativeInterfaceDetails{
		interfaceIndex: helperField("interfaceId", defaultIface), // yes interfaceId contains the index ("LAN1")
		interfaceID:    &defaultIface.Entity.ID,                  // entityID (1234)
		interfaceName:  helperField("interfaceName", defaultIface),
		destType:       helperField("destType", defaultIface),
	}
	return ifaceDetails
}

// assignInterfaceIndex attempts to assign the desired interface index for the native range by calling SiteExchangeSocketPorts API.
func (r *socketSiteResource) assignInterfaceIndex(ctx context.Context, currentInterfaceIndex cato_models.SocketInterfaceIDEnum,
	plan *tf.SocketSite, siteID string, diags *diag.Diagnostics,
) {
	var tfNativeRange tf.NativeRange
	if utils.CheckErr(diags, plan.NativeRange.As(ctx, &tfNativeRange, basetypes.ObjectAsOptions{})) {
		return
	}
	if !utils.HasValue(tfNativeRange.InterfaceIndex) {
		return
	}

	desiredInterfaceIndex := cato_models.SocketInterfaceIDEnum(tfNativeRange.InterfaceIndex.ValueString())

	if desiredInterfaceIndex == currentInterfaceIndex {
		return // No need to move the interface if the desired index is the default
	}

	exchangeInput := cato_models.ExchangeSocketPortsInput{
		Site:            &cato_models.SiteRefInput{By: cato_models.ObjectRefByID, Input: siteID},
		FirstInterface:  &cato_models.SocketInterfaceRefInput{InterfaceID: currentInterfaceIndex},
		SecondInterface: &cato_models.SocketInterfaceRefInput{InterfaceID: desiredInterfaceIndex},
	}

	siteExchangeSocketPortsResponse, err := r.client.catov2.SiteExchangeSocketPorts(ctx, r.client.AccountId, exchangeInput)
	if err != nil {
		diags.AddError("Catov2 API SiteExchangeSocketPorts error", err.Error())
		return
	}
	tflog.Debug(ctx, "attemptReassignNativeRangeIndex.tmpSiteUpdateSocketInterface.tmp.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteExchangeSocketPortsResponse),
	})
}

// updateNativeRange updates the native network range details by calling SiteUpdateNetworkRange API.
func (r *socketSiteResource) updateNetworkRange(ctx context.Context, plan *tf.SocketSite, networkRangeID string,
	isHA bool, diags *diag.Diagnostics,
) {
	networkRangeInput := r.prepareNetworkRangeInput(ctx, plan, isHA, diags)
	if diags.HasError() {
		return
	}

	// Update native network range
	_, err := r.client.catov2.SiteUpdateNetworkRange(ctx, networkRangeID, *networkRangeInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateNetworkRange error", err.Error())
	}
}

// updateSocketInterface updates the socket interface details.
func (r *socketSiteResource) updateSocketInterface(ctx context.Context, plan *tf.SocketSite, siteID string,
	isHA bool, diags *diag.Diagnostics,
) {
	if isHA && plan.ConnectionType.ValueString() == string(cato_models.SiteConnectionTypeEnumSocketGCP1500) {
		return // update does not work on GCP HA
	}

	socketIfaceInput, interfaceID := r.prepareSocketInterfaceInput(ctx, plan, isHA, diags)
	if diags.HasError() {
		return
	}

	// Update socket interface
	_, err := r.client.catov2.SiteUpdateSocketInterface(ctx, siteID, interfaceID, *socketIfaceInput, r.client.AccountId)
	if err != nil {
		diags.AddError("Catov2 API SiteUpdateSocketInterface error", err.Error())
	}
}

// isHA determines if the site is in HA scenario based on the number of sockets
func (r *socketSiteResource) isHA(state *tf.SocketSite) bool {
	if state == nil || !utils.HasValue(state.Sockets) {
		return false
	}
	return len(state.Sockets.Elements()) > 1
}
