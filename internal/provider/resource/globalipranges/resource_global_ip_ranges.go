package globalipranges

import (
	"cmp"
	"context"
	"slices"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	catoclient "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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
	client *catoclient.Service

	apiClient ResourceClient
}

func (r *globalIPRangesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_ip_ranges"
}

func (r *globalIPRangesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *globalIPRangesResource) ImportState(
	ctx context.Context,
	_ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
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
		rangeObj := convertCreateValues(ctx, diags, tfRange)
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
	planDetails := convertUpdateInput(
		ctx,
		r,
		cfg,
		state,
		&cfgRanges,
		&stateRanges,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}
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
		rangeObj := convertUpdateValues(ctx, diags, tfRange)
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

	convertDeleteInput(
		ctx,
		r,
		state,
		&stateRanges,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
}

// hydrate is used in Read and Import to get the current state of global IP ranges from the API and convert it to the TF model
func (r *globalIPRangesResource) hydrate(ctx context.Context, diags *diag.Diagnostics) *tf.GlobalIPRangesModel {
	result, err := r.resourceAPI().ObjectGlobalIPRangeList(ctx, r.client.AccountId, nil)
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
		rangeObj := convertHydrateValues2(ctx, diags, item)
		if diags.HasError() {
			return nil
		}
		globalIPRanges = append(globalIPRanges, rangeObj)
	}

	globalIPRangesSet := convertHydrateValues(diags, globalIPRanges)
	if diags.HasError() {
		return nil
	}

	return &tf.GlobalIPRangesModel{Ranges: globalIPRangesSet}
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	ctx context.Context,
	r *globalIPRangesResource,
	state *tf.GlobalIPRangesModel,
	originalStateRanges *[]tf.GlobalIPRange,
	diagnostics *diag.Diagnostics,
) {
	stateRanges := *originalStateRanges
	defer func() { *originalStateRanges = stateRanges }()

	if utils.CheckErr(diagnostics, state.Ranges.ElementsAs(ctx, &stateRanges, false)) {
		return
	}

	r.deleteRanges(ctx, stateRanges, diagnostics)
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

// updateRanges prepares the input and calls the Update API
func (r *globalIPRangesResource) updateRanges(ctx context.Context, toUpdate []tf.GlobalIPRange, diags *diag.Diagnostics,
) []tf.GlobalIPRange {
	if len(toUpdate) == 0 {
		return nil
	}

	input := r.prepareUpdateInput(toUpdate)
	return r.callUpdateRangesAPI(ctx, input, diags)
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
