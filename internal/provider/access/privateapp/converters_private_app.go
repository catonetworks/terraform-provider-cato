package privateapp

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	scalars "github.com/catonetworks/cato-go-sdk/scalars"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *privateAppResource) parseProtocolPorts(
	ctx context.Context,
	protoPorts []*privateAppProtocolPorts,
	diags *diag.Diagnostics,
) types.Set {
	var objDiags diag.Diagnostics
	setNull := types.SetNull(types.ObjectType{AttrTypes: ProtocolPortTypes})

	if protoPorts == nil {
		return setNull
	}

	protoPortsObjects := make([]types.Object, 0, len(protoPorts))

	for _, pp := range protoPorts {
		if pp == nil {
			continue
		}

		// Ports
		tfPortSet := types.SetNull(types.Int64Type)
		if pp.Port != nil {
			intSlice := make([]types.Int64, 0, len(pp.Port))
			for _, p := range pp.Port {
				intSlice = append(intSlice, types.Int64Value(p.GetInt64()))
			}
			tfPortSet, objDiags = types.SetValueFrom(ctx, types.Int64Type, intSlice)
			diags.Append(objDiags...)
			if diags.HasError() {
				return setNull
			}
		}

		// Port range
		tfPortRangeObj := types.ObjectNull(PortRangeTypes)
		if pp.PortRange != nil {
			tfPortRange := PortRange{
				From: types.Int64Value(pp.PortRange.From.GetInt64()),
				To:   types.Int64Value(pp.PortRange.To.GetInt64()),
			}
			tfPortRangeObj, objDiags = types.ObjectValueFrom(ctx, PortRangeTypes, tfPortRange)
			diags.Append(objDiags...)
			if diags.HasError() {
				return setNull
			}
		}

		// ProtocolPorts item
		tfPPort := ProtocolPort{
			Ports:     tfPortSet,
			PortRange: tfPortRangeObj,
			Protocol:  types.StringValue(string(pp.Protocol)),
		}
		tfPPortObj, objDiags := types.ObjectValueFrom(ctx, ProtocolPortTypes, tfPPort)
		diags.Append(objDiags...)
		if diags.HasError() {
			return setNull
		}
		protoPortsObjects = append(protoPortsObjects, tfPPortObj)
	}

	// convert slice to types.Set
	tfProtoPortSet, objDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ProtocolPortTypes}, protoPortsObjects)
	diags.Append(objDiags...)
	if diags.HasError() {
		return setNull
	}

	return tfProtoPortSet
}

func (r *privateAppResource) parsePublishedAppDomain(
	ctx context.Context,
	domain *privateAppPublishedAppDomain,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	objNull := types.ObjectNull(PublishedAppDomainTypes)

	if domain == nil {
		return objNull
	}

	tfDomain := PublishedAppDomain{
		ConnectorGroupName: types.StringPointerValue(domain.ConnectorGroupName),
		CreationTime:       types.StringValue(domain.CreationTime),
		ID:                 types.StringValue(domain.ID),
		PublishedAppDomain: types.StringValue(domain.PublishedAppDomain),
	}

	domainObj, objDiags := types.ObjectValueFrom(ctx, PublishedAppDomainTypes, tfDomain)
	diags.Append(objDiags...)
	if diags.HasError() {
		return objNull
	}
	return domainObj
}

func (r *privateAppResource) parsePrivateAppProbing(
	ctx context.Context,
	probing *privateAppProbing,
	diags *diag.Diagnostics,
) types.Object {
	var objDiags diag.Diagnostics
	objNull := types.ObjectNull(ProbingTypes)

	if probing == nil {
		return objNull
	}

	tfProbing := Probing{
		FaultThresholdDown: types.Int64Value(probing.FaultThresholdDown),
		ID:                 types.StringValue(probing.ID),
		Interval:           types.Int64Value(probing.Interval),
		Type:               types.StringValue(probing.Type),
	}

	probingObj, objDiags := types.ObjectValueFrom(ctx, ProbingTypes, tfProbing)
	diags.Append(objDiags...)
	if diags.HasError() {
		return objNull
	}
	return probingObj
}

func (r *privateAppResource) preparePublishedAppDomain(ctx context.Context, appDomain types.Object, diags *diag.Diagnostics,
) *cato_models.PublishedAppDomainInput {
	if !utils.HasValue(appDomain) {
		return nil
	}

	var tfAppDomain PublishedAppDomain
	if utils.CheckErr(diags, appDomain.As(ctx, &tfAppDomain, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.PublishedAppDomainInput{
		ConnectorGroupName: convert.KnownStringPointer(tfAppDomain.ConnectorGroupName),
		CreationTime:       convert.KnownStringPointer(tfAppDomain.CreationTime),
		ID:                 convert.KnownStringPointer(tfAppDomain.ID),
		PublishedAppDomain: convert.KnownStringPointer(tfAppDomain.PublishedAppDomain),
	}
}

func (r *privateAppResource) preparePrivateAppProbing(ctx context.Context, probing types.Object, diags *diag.Diagnostics,
) *cato_models.PrivateAppProbingInput {
	if !utils.HasValue(probing) {
		return nil
	}

	var tfProbing Probing
	if utils.CheckErr(diags, probing.As(ctx, &tfProbing, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.PrivateAppProbingInput{
		FaultThresholdDown: convert.KnownInt64Pointer(tfProbing.FaultThresholdDown),
		ID:                 convert.KnownStringPointer(tfProbing.ID),
		Interval:           convert.KnownInt64Pointer(tfProbing.Interval),
		Type:               convert.KnownStringPointer(tfProbing.Type),
	}
}

func (r *privateAppResource) prepareProtocolPorts(
	ctx context.Context,
	protPorts types.Set,
	diags *diag.Diagnostics,
) []*cato_models.CustomServiceInput {
	if !utils.HasValue(protPorts) {
		return nil
	}

	var out []*cato_models.CustomServiceInput

	for _, p := range protPorts.Elements() {
		var svcInput cato_models.CustomServiceInput

		if !utils.HasValue(p) {
			continue
		}
		port := p.(types.Object)

		var tfProtoPort ProtocolPort
		if utils.CheckErr(diags, port.As(ctx, &tfProtoPort, basetypes.ObjectAsOptions{})) {
			return nil
		}

		// Port numbers
		if utils.HasValue(tfProtoPort.Ports) {
			var tfPortNumbers []types.Int64
			if utils.CheckErr(diags, tfProtoPort.Ports.ElementsAs(ctx, &tfPortNumbers, false)) {
				return nil
			}
			for _, portNum := range tfPortNumbers {
				if utils.HasValue(portNum) {
					svcInput.Port = append(svcInput.Port, scalars.Port(portNum.String()))
				}
			}
		}

		// Port range
		if utils.HasValue(tfProtoPort.PortRange) {
			var tfProtoRange PortRange
			if utils.CheckErr(diags, tfProtoPort.PortRange.As(ctx, &tfProtoRange, basetypes.ObjectAsOptions{})) {
				return nil
			}
			svcInput.PortRange = &cato_models.PortRangeInput{
				From: scalars.Port(tfProtoRange.From.String()),
				To:   scalars.Port(tfProtoRange.To.String()),
			}
		}

		// Protocol
		svcInput.Protocol = cato_models.IPProtocol(tfProtoPort.Protocol.ValueString())

		out = append(out, &svcInput)
	}

	return out
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state Model) (resultInput cato_models.DeletePrivateApplicationInput) {
	resultInput = cato_models.DeletePrivateApplicationInput{
		PrivateApplication: &cato_models.PrivateApplicationRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.ID.ValueString(),
		},
	}
	return resultInput
}

// convertHydratePrivateAppStateValues converts between the Terraform model and SDK values.
func convertHydratePrivateAppStateValues(
	privateAppID string,
) (resultInput cato_models.PrivateApplicationRefInput) {
	resultInput = cato_models.PrivateApplicationRefInput{
		By:    cato_models.ObjectRefByID,
		Input: privateAppID,
	}
	return resultInput
}

func privateApplicationInputFromTerraform(
	ctx context.Context,
	r *privateAppResource,
	diagnostics *diag.Diagnostics,
	plan Model,
) cato_models.CreatePrivateApplicationInput {
	return cato_models.CreatePrivateApplicationInput{
		AllowICMPProtocol:  plan.AllowIcmpProtocol.ValueBool(),
		Description:        convert.KnownStringPointer(plan.Description),
		InternalAppAddress: plan.InternalAppAddress.ValueString(),
		Name:               plan.Name.ValueString(),
		PrivateAppProbing:  r.preparePrivateAppProbing(ctx, plan.PrivateAppProbing, diagnostics),
		ProbingEnabled:     plan.ProbingEnabled.ValueBool(),
		ProtocolPorts:      r.prepareProtocolPorts(ctx, plan.ProtocolPorts, diagnostics),
		Published:          plan.Published.ValueBool(),
		PublishedAppDomain: r.preparePublishedAppDomain(ctx, plan.PublishedAppDomain, diagnostics),
	}
}
