package group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}
var _ resource.ResourceWithModifyPlan = &GroupResource{}

type GroupResource struct {
	cfg *common.ProviderConfig
}

func NewResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// This description is used by the documentation generator and the language
	// server.
	resp.Schema.Description = "A resource which handles permission groups."
	resp.Schema.Attributes = cmlschema.Group()
	resp.Diagnostics = nil
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
