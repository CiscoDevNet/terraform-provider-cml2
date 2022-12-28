package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

func (r *LabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var (
		labModel cmlschema.LabModel
		err      error
	)

	tflog.Info(ctx, "Resource Lab CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &labModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lab := cmlclient.Lab{}
	if !labModel.Notes.IsNull() {
		lab.Notes = labModel.Notes.ValueString()
	}
	if !labModel.Description.IsNull() {
		lab.Description = labModel.Description.ValueString()
	}
	if !labModel.Title.IsNull() {
		lab.Title = labModel.Title.ValueString()
	}

	newLab, err := r.cfg.Client().LabCreate(ctx, lab)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to create lab, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewLab(ctx, newLab, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.LabAttrType},
			&labModel,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &labModel)...)

	tflog.Info(ctx, "Resource Lab CREATE done")
}
