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
		data *cmlschema.LabModel
		err  error
	)

	tflog.Info(ctx, "Resource Lab CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lab := cmlclient.Lab{}
	if !data.Notes.IsNull() {
		lab.Notes = data.Notes.ValueString()
	}
	if !data.Description.IsNull() {
		lab.Description = data.Description.ValueString()
	}
	if !data.Title.IsNull() {
		lab.Title = data.Title.ValueString()
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
			&data,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Lab CREATE: done")
}
