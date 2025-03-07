package lifecycle

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *LabLifecycleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data cmlschema.LabLifecycleModel

	tflog.Info(ctx, "Resource Lifecycle DELETE")

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Topology.IsNull() {
		tflog.Warn(ctx, "won't destroy as there's no topology")
		return
	}

	lab, err := r.cfg.Client().LabGet(ctx, data.LabID.ValueString(), false)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to read CML2 lab, got error: %s", err),
		)
		return
	}

	if lab.State != cmlclient.LabStateDefined {
		if lab.State == cmlclient.LabStateStarted {
			r.stop(ctx, resp.Diagnostics, data.LabID.ValueString())
		}
		r.wipe(ctx, resp.Diagnostics, data.LabID.ValueString())
	}

	err = r.cfg.Client().LabDestroy(ctx, data.LabID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "Resource Lifecycle DELETE done")
}
