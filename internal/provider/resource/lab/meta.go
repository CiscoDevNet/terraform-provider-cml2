package lab

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

const CML2ErrorLabel string = "resource lab"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &LabResource{}
var _ resource.ResourceWithImportState = &LabResource{}

type LabResource struct {
	cfg *common.ProviderConfig
}

func NewResource() resource.Resource {
	return &LabResource{}
}

func (r *LabResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *LabResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the
		// language server.
		Description: "A CML lab resource represents a CML lab. At create time, lab title, lab description and lab notes can be provided",
		Attributes:  schema.Lab(),
	}, nil
}

func (r *LabResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lab"
}

func (r LabResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
