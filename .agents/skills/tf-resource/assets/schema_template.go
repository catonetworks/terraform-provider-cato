package <package_name>


import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *<camelCaseResourceName>) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	_ = resp
}