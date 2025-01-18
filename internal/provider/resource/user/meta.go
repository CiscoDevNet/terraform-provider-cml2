package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

type UserResource struct {
	cfg *common.ProviderConfig
}

func NewResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// This description is used by the documentation generator and the language
	// server.
	resp.Schema.Description = "A resource which handles users."
	resp.Schema.Attributes = cmlschema.User()
	resp.Diagnostics = nil
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// resp.State.SetAttribute(ctx, path.Root("password"), )
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
