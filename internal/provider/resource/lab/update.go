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

func (r LabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var (
		data cmlschema.LabModel
		err  error
	)

	tflog.Info(ctx, "Resource Lab UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lab := cmlclient.Lab{
		ID:          data.ID.ValueString(),
		Notes:       data.Notes.ValueString(),
		Description: data.Description.ValueString(),
		Title:       data.Title.ValueString(),
	}

	newLab, err := r.cfg.Client().LabUpdate(ctx, lab)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to update lab, got error: %s", err),
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

	tflog.Info(ctx, "Resource Lab UPDATE done")
}
