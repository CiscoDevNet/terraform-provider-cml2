package link

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

const CML2ErrorLabel string = "CML resource link"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &LinkResource{}
var _ resource.ResourceWithImportState = &LinkResource{}
var _ resource.ResourceWithModifyPlan = &LinkResource{}

type LinkResource struct {
	cfg *common.ProviderConfig
	// client *cmlclient.Client
	// mu     *sync.Mutex
}

func NewResource() resource.Resource {
	return &LinkResource{}
}

func (r *LinkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *LinkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// This description is used by the documentation generator and the language
	// server.
	resp.Schema.Description = "A link resource represents a CML link. At create time, the lab ID, source and destination node ID are required.  Interface slots are optional.  By default, the next free interface slot is used."
	resp.Schema.Attributes = cmlschema.Link()
	resp.Diagnostics = nil
}

func (r *LinkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_link"
}

func (r LinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
