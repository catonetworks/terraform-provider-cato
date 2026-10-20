package networkrange

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

func hydrateAddNetworkRangeAPI(
	ctx context.Context,
	client dhcp.SdkClienter,
	cfg, plan *tf.NetworkRange,
	diags *diag.Diagnostics,
) cato_models.AddNetworkRangeInput {
	rangeType := cato_models.SubnetType(plan.RangeType.ValueString())

	return cato_models.AddNetworkRangeInput{
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, client, rangeType, plan.DhcpSettings, diags),
		Gateway:          convert.KnownStringPointer(plan.Gateway),
		InternetOnly:     convert.KnownBoolPointer(plan.InternetOnly),
		LocalIP:          convert.KnownStringPointer(plan.LocalIP),
		MdnsReflector:    convert.KnownBoolPointer(plan.MdnsReflector),
		Name:             plan.Name.ValueString(),
		RangeType:        rangeType,
		Subnet:           plan.Subnet.ValueString(),
		TranslatedSubnet: convert.TranslatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
		Vlan:             convert.KnownInt64Pointer(plan.Vlan),
	}
}

func hydrateUpdateNetworkRangeAPI(
	ctx context.Context,
	client dhcp.SdkClienter,
	cfg, plan *tf.NetworkRange,
	diags *diag.Diagnostics,
) cato_models.UpdateNetworkRangeInput {
	rangeType := cato_models.SubnetType(plan.RangeType.ValueString())

	return cato_models.UpdateNetworkRangeInput{
		DhcpSettings:     dhcp.PrepareDHCPSettings(ctx, client, rangeType, plan.DhcpSettings, diags),
		Gateway:          convert.KnownStringPointer(plan.Gateway),
		InternetOnly:     convert.KnownBoolPointer(plan.InternetOnly),
		LocalIP:          convert.KnownStringPointer(plan.LocalIP),
		MdnsReflector:    convert.KnownBoolPointer(plan.MdnsReflector),
		Name:             convert.KnownStringPointer(plan.Name),
		RangeType:        (*cato_models.SubnetType)(convert.KnownStringPointer(plan.RangeType)),
		Subnet:           convert.KnownStringPointer(plan.Subnet),
		TranslatedSubnet: convert.TranslatedSubnetForAPIInput(cfg.TranslatedSubnet, plan.TranslatedSubnet),
		Vlan:             convert.KnownInt64Pointer(plan.Vlan),
	}
}
