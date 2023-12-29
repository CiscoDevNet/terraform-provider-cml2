package lab

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &LabResource{}
	_ resource.ResourceWithImportState = &LabResource{}
	_ resource.ResourceWithModifyPlan  = &LabResource{}
)

type LabResource struct {
	cfg *common.ProviderConfig
}

func NewResource() resource.Resource {
	return &LabResource{}
}

func (r *LabResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *LabResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// This description is used by the documentation generator and the language
	// server.
	resp.Schema.Description = "A lab resource represents a CML lab. At create time, a lab title, lab description and lab notes can be provided."
	resp.Schema.Attributes = cmlschema.Lab()
	resp.Diagnostics = nil
}

func (r *LabResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lab"
}

func (r LabResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
