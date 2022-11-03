package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r LabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var (
		planData schema.LabModel
		err      error
	)

	tflog.Info(ctx, "Resource Lab UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lab := cmlclient.Lab{
		ID:          planData.ID.ValueString(),
		Notes:       planData.Notes.ValueString(),
		Description: planData.Description.ValueString(),
		Title:       planData.Title.ValueString(),
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
			schema.NewLab(ctx, newLab, &resp.Diagnostics),
			types.ObjectType{AttrTypes: schema.LabAttrType},
			&planData,
		)...,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Lab UPDATE: done")
}
