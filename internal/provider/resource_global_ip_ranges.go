package provider

import (
	"cmp"
	"context"
	"slices"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

var (
	_ resource.Resource                = &globalIPRangesResource{}
	_ resource.ResourceWithConfigure   = &globalIPRangesResource{}
	_ resource.ResourceWithImportState = &globalIPRangesResource{}
)

func NewGlobalIPRangesResource() resource.Resource {
	return &globalIPRangesResource{}
}

type globalIPRangesResource struct {
	client *catoClientData
}

type ipRangePlanDetails struct {
	toCreate []tf.GlobalIPRange
	toUpdate []tf.GlobalIPRange
	toDelete []tf.GlobalIPRange
	toKeep   []tf.GlobalIPRange
}

func (r *globalIPRangesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_ip_ranges"
}

func (r *globalIPRangesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_global_ip_ranges` resource contains the configuration parameters for " +
			"global IP ranges in the Cato platform. " +
			"There should be only one resource of this type in the TF config and it manages all the global IP ranges for a given account.",

		Attributes: map[string]schema.Attribute{
			"ranges": schema.SetNestedAttribute{
				Description: "List of global IP ranges",
				Required:    true,
				Validators:  []validator.Set{validators.GetGlobalIPRangeValidator()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Description: "Global IP range description",
							Optional:    true,
						},
						"id": schema.StringAttribute{
							Description: "Global IP range ID",
							Computed:    true,
							// PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"ip_range": schema.StringAttribute{
							Description: "Global IP range",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Global IP range name",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// ModifyPlan computes what ranges should be updated, created or deleted based on the config and state.
func (r *globalIPRangesResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var cfg *tf.GlobalIPRangesModel
	var cfgRanges []tf.GlobalIPRange
	var stateRanges []tf.GlobalIPRange
	state := &tf.GlobalIPRangesModel{} // avoid nil pointer dereference
	stateDefined := !req.State.Raw.IsNull()

	if req.Plan.Raw.IsNull() { // resource destruction
		return
	}

	// get config, validate it
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if cfg == nil || !utils.HasValue(cfg.Ranges) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, cfg.Ranges.ElementsAs(ctx, &cfgRanges, false)) {
		return
	}
	ipRangeValidator := validators.GetGlobalIPRangeValidator()
	ipRangeValidator.ValidateGlobalIPRange(cfgRanges, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// get state
	if stateDefined {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if state != nil && utils.HasValue(state.Ranges) {
			if utils.CheckErr(&resp.Diagnostics, state.Ranges.ElementsAs(ctx, &stateRanges, false)) {
				return
			}
		}
	}

	// prepare plan based on state and config
	plan, _ := r.computePlan(ctx, stateRanges, cfgRanges, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *globalIPRangesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *globalIPRangesResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if state := r.hydrate(ctx, &resp.Diagnostics); state != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

// Create creates global IP ranges in bulk.
func (r *globalIPRangesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tf.GlobalIPRangesModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ranges []tf.GlobalIPRange
	resp.Diagnostics.Append(plan.Ranges.ElementsAs(ctx, &ranges, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// call create API
	input := r.prepareCreateInput(ranges)
	newRanges := r.callCreateRangesAPI(ctx, input, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// prepare new state (with IDs from API response)
	globalIPRanges := make([]attr.Value, 0, len(newRanges))
	for _, tfRange := range newRanges {
		rangeObj, objDiags := types.ObjectValueFrom(ctx, tf.GlobalIPRangeTypes, tfRange)
		diags.Append(objDiags...)
		if diags.HasError() {
			return
		}
		globalIPRanges = append(globalIPRanges, rangeObj)
	}
	globalIPRangesSet, setDiags := types.SetValue(types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}, globalIPRanges)
	diags.Append(setDiags...)
	if diags.HasError() {
		return
	}
	newState := tf.GlobalIPRangesModel{Ranges: globalIPRangesSet}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Read the global IP ranges, this is needed to refresh the state after create/update and for import
func (r *globalIPRangesResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state := r.hydrate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update the IP ranges
// based on the config and state, figure out which ranges need to be created, updated or deleted and call the respective APIs
func (r *globalIPRangesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var cfg, state *tf.GlobalIPRangesModel
	var cfgRanges, stateRanges []tf.GlobalIPRange

	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// prepare a plan based on state and config, this will determine which ranges need to be created, updated or deleted
	if utils.CheckErr(&resp.Diagnostics, state.Ranges.ElementsAs(ctx, &stateRanges, false)) {
		return
	}
	if utils.CheckErr(&resp.Diagnostics, cfg.Ranges.ElementsAs(ctx, &cfgRanges, false)) {
		return
	}
	_, planDetails := r.computePlan(ctx, stateRanges, cfgRanges, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// call respective APIs (if needed) to delete, update or create ranges based on the planDetails computed above
	r.deleteRanges(ctx, planDetails.toDelete, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated := r.updateRanges(ctx, planDetails.toUpdate, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	created := r.createNewRanges(ctx, planDetails.toCreate, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// prepare new state with the results from API calls
	newRanges := created
	newRanges = append(newRanges, updated...)
	newRanges = append(newRanges, planDetails.toKeep...)
	slices.SortFunc(newRanges, func(i, j tf.GlobalIPRange) int {
		return cmp.Compare(i.ID.ValueString(), j.ID.ValueString())
	})

	globalIPRanges := make([]attr.Value, 0, len(newRanges))
	for _, tfRange := range newRanges {
		rangeObj, objDiags := types.ObjectValueFrom(ctx, tf.GlobalIPRangeTypes, tfRange)
		diags.Append(objDiags...)
		if diags.HasError() {
			return
		}
		globalIPRanges = append(globalIPRanges, rangeObj)
	}
	globalIPRangesSet, setDiags := types.SetValue(types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}, globalIPRanges)
	diags.Append(setDiags...)
	if diags.HasError() {
		return
	}

	// set the new state
	newState := tf.GlobalIPRangesModel{Ranges: globalIPRangesSet}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Delete all the global IP ranges in the state
func (r *globalIPRangesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *tf.GlobalIPRangesModel
	var stateRanges []tf.GlobalIPRange

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if utils.CheckErr(&resp.Diagnostics, state.Ranges.ElementsAs(ctx, &stateRanges, false)) {
		return
	}

	r.deleteRanges(ctx, stateRanges, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
}

// hydrate is used in Read and Import to get the current state of global IP ranges from the API and convert it to the TF model
func (r *globalIPRangesResource) hydrate(ctx context.Context, diags *diag.Diagnostics) *tf.GlobalIPRangesModel {
	result, err := r.client.catov2.ObjectGlobalIPRangeList(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("Error fetching global IP ranges", err.Error())
		return nil
	}

	items := result.GetObject().GetGlobalIPRangeList().GetItems()
	slices.SortFunc(items, func(i, j *cato_go_sdk.ObjectGlobalIpRangeList_Object_GlobalIPRangeList_Items) int {
		return cmp.Compare(i.GetID(), j.GetID())
	})
	globalIPRanges := make([]attr.Value, 0, len(items))
	for _, item := range items {
		tfRange := tf.GlobalIPRange{
			Description: types.StringPointerValue(item.GetDescription()),
			ID:          types.StringValue(item.GetID()),
			IPRange:     types.StringValue(item.GetIPRange()),
			Name:        types.StringValue(item.GetName()),
		}
		rangeObj, objDiags := types.ObjectValueFrom(ctx, tf.GlobalIPRangeTypes, tfRange)
		diags.Append(objDiags...)
		if diags.HasError() {
			return nil
		}
		globalIPRanges = append(globalIPRanges, rangeObj)
	}

	globalIPRangesSet, setDiags := types.SetValue(types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}, globalIPRanges)
	diags.Append(setDiags...)
	if diags.HasError() {
		return nil
	}

	return &tf.GlobalIPRangesModel{Ranges: globalIPRangesSet}
}

// computePlan compares the config and state and determines which ranges need to be created, updated, deleted or kept as is.
func (r *globalIPRangesResource) computePlan(ctx context.Context, state, config []tf.GlobalIPRange, diags *diag.Diagnostics,
) (newPlan *tf.GlobalIPRangesModel, planDetails ipRangePlanDetails) {
	var plan []tf.GlobalIPRange

	// map of state ranges by name
	stateMap := make(map[string]*tf.GlobalIPRange)
	for i, item := range state {
		stateMap[item.Name.ValueString()] = &state[i]
	}

	// iterate config, if found in state, use ID from state, otherwise ID is unknown (new resource)
	for _, cfgItem := range config {
		planItem := cfgItem
		if planItem.Description.IsNull() {
			planItem.Description = types.StringValue("") // API always returns empty string
		}

		planItem.ID = types.StringUnknown() // ID defaults to unknown, if not found in the state
		if stateItem, ok := stateMap[cfgItem.Name.ValueString()]; ok {
			planItem.ID = stateItem.ID // ID is from state
			if planItem.Equal(*stateItem) {
				planDetails.toKeep = append(planDetails.toKeep, planItem)
			} else {
				planDetails.toUpdate = append(planDetails.toUpdate, planItem)
			}
			delete(stateMap, cfgItem.Name.ValueString())
		} else {
			planDetails.toCreate = append(planDetails.toCreate, planItem)
		}
		plan = append(plan, planItem)
	}

	// items in the state no longer in config should be deleted
	for _, stateItem := range stateMap {
		planDetails.toDelete = append(planDetails.toDelete, *stateItem)
	}

	tfSet, setDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: tf.GlobalIPRangeTypes}, plan)
	diags.Append(setDiags...)
	if diags.HasError() {
		return nil, planDetails
	}
	return &tf.GlobalIPRangesModel{Ranges: tfSet}, planDetails
}

// prepareCreateInput converts the list of ranges to be created to the input format required by the Create API
func (r *globalIPRangesResource) prepareCreateInput(ranges []tf.GlobalIPRange) []*cato_models.CreateGlobalIPRangeInput {
	input := make([]*cato_models.CreateGlobalIPRangeInput, 0, len(ranges))
	for _, globalIPRange := range ranges {
		inputItem := cato_models.CreateGlobalIPRangeInput{
			IPRange: globalIPRange.IPRange.ValueString(),
			Name:    globalIPRange.Name.ValueString(),
		}
		if utils.HasValue(globalIPRange.Description) {
			inputItem.Description = globalIPRange.Description.ValueStringPointer()
		}
		input = append(input, &inputItem)
	}
	return input
}

// prepareUpdateInput converts the list of ranges to be updated to the input format required by the Update API
func (r *globalIPRangesResource) prepareUpdateInput(ranges []tf.GlobalIPRange) []*cato_models.UpdateGlobalIPRangeInput {
	input := make([]*cato_models.UpdateGlobalIPRangeInput, 0, len(ranges))
	for _, globalIPRange := range ranges {
		inputItem := cato_models.UpdateGlobalIPRangeInput{
			ID:      globalIPRange.ID.ValueString(),
			IPRange: globalIPRange.IPRange.ValueStringPointer(),
			Name:    globalIPRange.Name.ValueStringPointer(),
		}
		if utils.HasValue(globalIPRange.Description) {
			inputItem.Description = globalIPRange.Description.ValueStringPointer()
		}
		input = append(input, &inputItem)
	}
	return input
}

// prepareDeleteInput converts the list of ranges to be deleted to the input format required by the Delete API
func (r *globalIPRangesResource) prepareDeleteInput(ranges []tf.GlobalIPRange) []*cato_models.GlobalIPRangeRefInput {
	input := make([]*cato_models.GlobalIPRangeRefInput, 0, len(ranges))
	for _, globalIPRange := range ranges {
		inputItem := cato_models.GlobalIPRangeRefInput{
			By:    cato_models.ObjectRefByID,
			Input: globalIPRange.ID.ValueString(),
		}
		input = append(input, &inputItem)
	}
	return input
}

// createNewRanges prepares the input andcalls the Create API
func (r *globalIPRangesResource) createNewRanges(ctx context.Context, toCreate []tf.GlobalIPRange, diags *diag.Diagnostics,
) []tf.GlobalIPRange {
	if len(toCreate) == 0 {
		return nil
	}

	input := r.prepareCreateInput(toCreate)
	return r.callCreateRangesAPI(ctx, input, diags)
}

// updateRanges prepares the input and calls the Update API
func (r *globalIPRangesResource) updateRanges(ctx context.Context, toUpdate []tf.GlobalIPRange, diags *diag.Diagnostics,
) []tf.GlobalIPRange {
	if len(toUpdate) == 0 {
		return nil
	}

	input := r.prepareUpdateInput(toUpdate)
	return r.callUpdateRangesAPI(ctx, input, diags)
}

// deleteRanges prepares the input and calls the Delete API
func (r *globalIPRangesResource) deleteRanges(ctx context.Context, toDelete []tf.GlobalIPRange, diags *diag.Diagnostics,
) {
	if len(toDelete) == 0 {
		return
	}

	input := r.prepareDeleteInput(toDelete)
	r.callDeleteRangesAPI(ctx, input, diags)
}

// callCreateRangesAPI calls the Create API and converts the response to the TF model
func (r *globalIPRangesResource) callCreateRangesAPI(ctx context.Context, input []*cato_models.CreateGlobalIPRangeInput,
	diags *diag.Diagnostics,
) (createdRanges []tf.GlobalIPRange) {
	result, err := r.client.catov2.ObjectCreateGlobalIPRangeBulk(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Error creating global IP ranges", err.Error())
		return nil
	}

	items := result.GetObject().GetCreateGlobalIPRangeBulk().GetGlobalIPRange()
	createdRanges = make([]tf.GlobalIPRange, 0, len(items))
	for _, item := range items {
		tfRange := tf.GlobalIPRange{
			Description: types.StringPointerValue(item.GetDescription()),
			ID:          types.StringValue(item.GetID()),
			IPRange:     types.StringValue(item.GetIPRange()),
			Name:        types.StringValue(item.GetName()),
		}
		createdRanges = append(createdRanges, tfRange)
	}
	return createdRanges
}

// callUpdateRangesAPI calls the Update API and converts the response to the TF model
func (r *globalIPRangesResource) callUpdateRangesAPI(ctx context.Context, input []*cato_models.UpdateGlobalIPRangeInput,
	diags *diag.Diagnostics,
) (updatedRanges []tf.GlobalIPRange) {
	result, err := r.client.catov2.ObjectUpdateGlobalIPRangeBulk(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Error updating global IP ranges", err.Error())
		return nil
	}

	items := result.GetObject().GetUpdateGlobalIPRangeBulk().GetGlobalIPRange()
	updatedRanges = make([]tf.GlobalIPRange, 0, len(items))
	for _, item := range items {
		tfRange := tf.GlobalIPRange{
			Description: types.StringPointerValue(item.GetDescription()),
			ID:          types.StringValue(item.GetID()),
			IPRange:     types.StringValue(item.GetIPRange()),
			Name:        types.StringValue(item.GetName()),
		}
		updatedRanges = append(updatedRanges, tfRange)
	}
	return updatedRanges
}

// callDeleteRangesAPI calls the Delete API
func (r *globalIPRangesResource) callDeleteRangesAPI(ctx context.Context, input []*cato_models.GlobalIPRangeRefInput,
	diags *diag.Diagnostics,
) {
	_, err := r.client.catov2.ObjectDeleteGlobalIPRangeBulk(ctx, r.client.AccountId, input)
	if err != nil {
		diags.AddError("Error deleting global IP ranges", err.Error())
		return
	}
}
