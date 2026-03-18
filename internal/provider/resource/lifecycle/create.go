package lifecycle

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Create creates (imports) and optionally starts a lab based on the configured topology.
func (r *LabLifecycleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		data cmlschema.LabLifecycleModel
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
		imported, importErr := r.cfg.Client().Lab.Import(ctx, data.Topology.ValueString())
		start.lab = &imported
		if importErr != nil {
			resp.Diagnostics.AddError(
				common.ErrorLabel,
				fmt.Sprintf("Unable to import lab, got error: %s", importErr),
			)
			return
		}
	} else {
		lab, getErr := r.cfg.Client().Lab.GetByID(ctx, models.UUID(data.LabID.ValueString()), true)
		start.lab = &lab
		if getErr != nil {
			resp.Diagnostics.AddError(
				common.ErrorLabel,
				fmt.Sprintf("Unable to get lab, got error: %s", getErr),
			)
			return
		}
	}

	// inject the configurations into the nodes
	r.injectConfigs(ctx, start.lab, &data, &resp.Diagnostics)

	// if unknown state or specifically "start" state, start the lab...
	// but only if there were no errors from config injection
	if !resp.Diagnostics.HasError() &&
		(data.State.IsUnknown() ||
			data.State.ValueString() == string(models.LabStateStarted)) {
		r.startNodes(ctx, &resp.Diagnostics, start)
	}

	// fetch lab again, with nodes and interfaces
	lab, err := r.cfg.Client().Lab.GetByID(ctx, start.lab.ID, true)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	data.LabID = types.StringValue(string(lab.ID))
	data.State = types.StringValue(string(lab.State))
	data.Nodes = r.populateNodes(ctx, &lab, &resp.Diagnostics)
	data.Booted = types.BoolValue(lab.Booted())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Resource Lifecycle CREATE done")
}
