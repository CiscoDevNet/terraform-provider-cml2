package lifecycle

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r *LabLifecycleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		data schema.LabLifecycleModel
		err  error
	)

	tflog.Info(ctx, "Resource Lifecycle CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a resource identifier
	if data.ID.IsUnknown() {
		data.ID = types.StringValue(uuid.New().String())
	}

	start := startData{
		staging:  getStaging(ctx, req.Config, &resp.Diagnostics),
		timeouts: getTimeouts(ctx, req.Config, &resp.Diagnostics),
		wait:     data.Wait.IsNull() || data.Wait.ValueBool(),
	}

	if data.LabID.IsUnknown() {
		tflog.Info(ctx, "Create: import")
		start.lab, err = r.cfg.Client().LabImport(ctx, data.Topology.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to import lab, got error: %s", err),
			)
			return
		}
		data.LabID = types.StringValue(start.lab.ID)
	} else {
		start.lab, err = r.cfg.Client().LabGet(ctx, data.LabID.ValueString(), true)
		if err != nil {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to get lab, got error: %s", err),
			)
			return
		}
	}

	// inject the configurations into the nodes
	r.injectConfigs(ctx, start.lab, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// if unknown state or specifically "start" state, start the lab...
	if data.State.IsUnknown() || data.State.ValueString() == cmlclient.LabStateStarted {
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

	data.LabID = types.StringValue(lab.ID)
	data.State = types.StringValue(lab.State)
	data.Nodes = r.populateNodes(ctx, lab, &resp.Diagnostics)
	data.Booted = types.BoolValue(lab.Booted())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Resource Lifecycle CREATE: done")
}
