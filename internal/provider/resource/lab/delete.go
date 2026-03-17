package lab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *LabResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		data cmlschema.LabModel
		err  error
	)

	tflog.Info(ctx, "Resource Lab DELETE")

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	common.Converge(
		ctx, r.cfg.Client(), &resp.Diagnostics,
		data.ID.ValueString(), "1h",
	)
	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := r.cfg.Client().Lab.GetByID(ctx, models.UUID(data.ID.ValueString()), false)
	if err != nil {
		resp.Diagnostics.AddWarning(
			common.ErrorLabel,
			fmt.Sprintf("Unable to read CML2 lab, got error: %s", err),
		)
		return
	}

	if lab.Running() {
		err = r.cfg.Client().Lab.Stop(ctx, models.UUID(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddWarning(
				common.ErrorLabel,
				fmt.Sprintf("Unable to stop CML2 lab, got error: %s", err),
			)
			return
		}
		err = r.cfg.Client().Lab.Wipe(ctx, models.UUID(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddWarning(
				common.ErrorLabel,
				fmt.Sprintf("Unable to wipe CML2 lab, got error: %s", err),
			)
			return
		}
	}

	if lab.State != models.LabStateDefined {
		resp.Diagnostics.AddError(common.ErrorLabel, "lab is not in DEFINED_ON_CORE state")
		return
	}

	err = r.cfg.Client().Lab.Delete(ctx, models.UUID(data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to destroy lab, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Resource Lab DELETE done")
}
