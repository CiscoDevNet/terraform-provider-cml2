// Package annotation implements the CML2 classic annotation resource.
package annotation

import (
	"context"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &AnnotationResource{}
	_ resource.ResourceWithImportState = &AnnotationResource{}
)

type AnnotationResource struct {
	cfg *common.ProviderConfig
}

func NewResource() resource.Resource {
	return &AnnotationResource{}
}

func (r *AnnotationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cfg = common.ResourceConfigure(ctx, req, resp)
}

func (r *AnnotationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotation"
}

func (r *AnnotationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.Description = "A classic annotation in a CML lab (currently: text annotations)."
	resp.Schema.Attributes = cmlschema.Annotation()
}

func (r AnnotationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: <lab_id>/<annotation_id>
	parts := common.Split2(req.ID, "/")
	if parts == nil {
		resp.Diagnostics.AddError(common.ErrorLabel, "invalid import id, expected <lab_id>/<annotation_id>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lab_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
