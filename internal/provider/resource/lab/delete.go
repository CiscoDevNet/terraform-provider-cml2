package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/common"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r *LabResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var (
		data *schema.LabModel
		err  error
	)

	tflog.Info(ctx, "Resource Lab DELETE")

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	common.Converge(ctx, r.cfg.Client(), &resp.Diagnostics, data.ID.Value, "1h")
	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := r.cfg.Client().LabGet(ctx, data.ID.Value, false)
	if err != nil {
		resp.Diagnostics.AddWarning(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to read CML2 lab, got error: %s", err),
		)
		return
	}

	if lab.Running() {
		err = r.cfg.Client().LabStop(ctx, data.ID.Value)
		if err != nil {
			resp.Diagnostics.AddWarning(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to stop CML2 lab, got error: %s", err),
			)
			return
		}
		err = r.cfg.Client().LabWipe(ctx, data.ID.Value)
		if err != nil {
			resp.Diagnostics.AddWarning(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to wipe CML2 lab, got error: %s", err),
			)
			return
		}
	}

	if lab.State != cmlclient.LabStateDefined {
		resp.Diagnostics.AddError(CML2ErrorLabel, "lab is not in DEFINED_ON_CORE state")
		return
	}

	err = r.cfg.Client().LabDestroy(ctx, data.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy lab, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Resource Lab DELETE: done")
}
