package node

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

const CML2ErrorLabel string = "CML resource node"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &NodeResource{}
var _ resource.ResourceWithImportState = &NodeResource{}

type NodeResource struct {
	cfg *common.ProviderConfig
}

func NewResource() resource.Resource {
	return &NodeResource{}
}

func (r *NodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *NodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// This description is used by the documentation generator and the language
	// server.
	resp.Schema.MarkdownDescription = "A node resource represents a CML node. At create time, the lab ID, a node definition and a label must be provided.  Other attributes are optional.  Note that some attributes can't be changed after the node state has changed to `STARTED` (see the lifecyle resource) once. Changing attributes will then require a replace."
	resp.Schema.Attributes = cmlschema.Node()
	resp.Diagnostics = nil
}

func (r *NodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (r NodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
