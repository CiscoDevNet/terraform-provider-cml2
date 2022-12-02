package lifecycle

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

func (r LabLifecycleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		configData, planData, stateData *cmlschema.LabLifecycleModel
		err                             error
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if stateData.State.ValueString() != planData.State.ValueString() {
		tflog.Info(ctx, "state changed")

		start := startData{
			staging:  getStaging(ctx, req.Config, &resp.Diagnostics),
			timeouts: getTimeouts(ctx, req.Config, &resp.Diagnostics),
			wait:     planData.Wait.IsNull() || planData.Wait.ValueBool(),
		}

		// need to get the lab data here
		start.lab, err = r.cfg.Client().LabGet(ctx, planData.LabID.ValueString(), true)
		if err != nil {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to fetch lab, got error: %s", err),
			)
			return
		}

		// this is very blunt ...
		if stateData.State.ValueString() == cmlclient.LabStateStarted {
			if planData.State.ValueString() == cmlclient.LabStateStopped {
				r.stop(ctx, resp.Diagnostics, planData.LabID.ValueString())
			}
			if planData.State.ValueString() == cmlclient.LabStateDefined {
				r.stop(ctx, resp.Diagnostics, planData.LabID.ValueString())
				timeout := start.timeouts.Update.ValueString()
				common.Converge(
					ctx, r.cfg.Client(), &resp.Diagnostics,
					planData.LabID.ValueString(), timeout,
				)
				r.wipe(ctx, resp.Diagnostics, planData.LabID.ValueString())
			}
		}

		if stateData.State.ValueString() == cmlclient.LabStateStopped {
			if planData.State.ValueString() == cmlclient.LabStateStarted {
				r.startNodes(ctx, &resp.Diagnostics, start)
			}
			if planData.State.ValueString() == cmlclient.LabStateDefined {
				r.wipe(ctx, resp.Diagnostics, planData.LabID.ValueString())
			}
		}

		if stateData.State.ValueString() == cmlclient.LabStateDefined {
			if planData.State.ValueString() == cmlclient.LabStateStarted {
				r.startNodes(ctx, &resp.Diagnostics, start)
			}
		}
		// not sure if this makes sense... state could change when not waiting
		// for convergence.  then again, there's no differentiation at the lab
		// level between "STARTED" and "BOOTED" (e.g. converged).  It's always
		// started...
		if start.wait {
			timeout := start.timeouts.Update.ValueString()
			common.Converge(
				ctx, r.cfg.Client(), &resp.Diagnostics,
				planData.LabID.ValueString(), timeout,
			)
		}
	}

	// since we have changed lab state, we need to re-read all the node
	// state...
	lab, err := r.cfg.Client().LabGet(ctx, planData.LabID.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to fetch lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Update: lab state: %s", lab.State))
	planData.State = types.StringValue(lab.State)
	planData.Nodes = r.populateNodes(ctx, lab, &resp.Diagnostics)
	planData.Booted = types.BoolValue(lab.Booted())

	resp.Diagnostics.Append(resp.State.Set(ctx, planData)...)
	tflog.Info(ctx, "Update: done")
}
