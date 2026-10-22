package <package_name>


import (
	"context"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &<camelCaseResourceName>{}
	_ resource.ResourceWithConfigure   = &<camelCaseResourceName>{}
	_ resource.ResourceWithImportState = &<camelCaseResourceName>{}
)

func New<PascalCaseResourceName>Resource() resource.Resource {
	return &<camelCaseResourceName>{}
}

type <camelCaseResourceName> struct {
	client     *catoclient.Service
	mockClient Client
}

type Client interface { }

func (r *<camelCaseResourceName>) getClient() Client {
	if r.mockClient != nil {
		return r.mockClient
	}
	if r.client == nil {
		return nil
	}
	return r.client.Catov2
}

func (r *<camelCaseResourceName>) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_<snail_case_resource_name_without_cato_prefix>"
}


func (r *<camelCaseResourceName>) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *<camelCaseResourceName>) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	_ = ctx
	_ = req
	_ = resp
}

func (r *<camelCaseResourceName>) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	_ = ctx
	_ = req
	_ = resp
}

func (r *<camelCaseResourceName>) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	_ = ctx
	_ = req
	_ = resp
}

func (r *<camelCaseResourceName>) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	_ = ctx
	_ = req
	_ = resp
}

func (r *<camelCaseResourceName>) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	_ = ctx
	_ = req
	_ = resp
}
