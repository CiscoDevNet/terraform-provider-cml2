// Package node implements the CML2 node resource.
package node

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &NodeResource{}
	_ resource.ResourceWithImportState = &NodeResource{}
	_ resource.ResourceWithModifyPlan  = &NodeResource{}
)

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
	resp.Schema.MarkdownDescription = "A node resource represents a CML node. At create time, the lab ID, a " +
		"node definition and a label must be provided.  Other attributes are optional.  Note that some " +
		"attributes can't be changed after the node state has changed to `STARTED` (see the `lifecyle` resource) " +
		"once. Changing attributes will then require a replace.  " +
		"Node configurations are \"day zero\" configurations. Replacing a configuration typically requires a " +
		"node replacement if the node has been started.  No Configurations can be provided for unmanaged switches. " +
		"External connectors require the connector label (like \"NAT\"), not the device name (like \"virbr0\"). " +
		"The available connectors can be retrieved via the external connector data source."
	resp.Schema.Attributes = cmlschema.Node()
	resp.Diagnostics = nil
}

func (r *NodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (r NodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
