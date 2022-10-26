package lifecycle

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r *LabLifecycleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		data *schema.LabLifecycleModel
		err  error
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	start := startData{
		staging:  getStaging(ctx, req.Config, &resp.Diagnostics),
		timeouts: getTimeouts(ctx, req.Config, &resp.Diagnostics),
		wait:     data.Wait.Null || data.Wait.Value,
	}

	if data.ID.IsUnknown() {
		tflog.Info(ctx, "Create: import")
		start.lab, err = r.cfg.Client().LabImport(ctx, data.Topology.Value)
		if err != nil {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to import lab, got error: %s", err),
			)
			return
		}
		data.ID = types.String{Value: start.lab.ID}
	} else {
		start.lab, err = r.cfg.Client().LabGet(ctx, data.ID.Value, true)
		if err != nil {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to get lab, got error: %s", err),
			)
			return
		}
	}

	// inject the configurations into the nodes
	r.injectConfigs(ctx, start.lab, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// if unknown state or specifically "start" state, start the lab...
	if data.State.Unknown || data.State.Value == cmlclient.LabStateStarted {
		r.startNodes(ctx, &resp.Diagnostics, start)
	}

	// fetch lab again, with nodes and interfaces
	lab, err := r.cfg.Client().LabGet(ctx, start.lab.ID, true)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	data.ID = types.String{Value: lab.ID}
	data.State = types.String{Value: lab.State}
	data.Nodes = r.populateNodes(ctx, lab, &resp.Diagnostics)
	data.Booted = types.Bool{Value: lab.Booted()}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	tflog.Info(ctx, "Create: done")
}
