package lifecycle

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// Update applies lifecycle state changes (start/stop/wipe) and refreshes state.
func (r LabLifecycleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		configData, planData, stateData cmlschema.LabLifecycleModel
		err                             error
	)

	tflog.Info(ctx, "Resource LabLifecycle UPDATE")

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

	desired := models.LabState(planData.State.ValueString())
	stateChanged := models.LabState(stateData.State.ValueString()) != desired
	wait := planData.Wait.IsNull() || planData.Wait.ValueBool()

	// Fetch current lab state once.  We need it both to detect drift (when
	// lifecycle.state is unchanged) and to drive the corrective action.
	lab, err := r.cfg.Client().Lab.GetByID(ctx, models.UUID(planData.LabID.ValueString()), true)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to fetch lab, got error: %s", err),
		)
		return
	}

	// Decide whether to act:
	// - Explicit lifecycle.state transition, OR
	// - Dependency drift (node/link state diverged while lifecycle.state stayed
	//   the same), OR
	// - Desired state is STARTED: always attempt startNodes because a dependent
	//   resource (e.g. an external_connector node) may have been replaced during
	//   this apply cycle.  Lab.Start / Node.Start are idempotent — already-running
	//   nodes are left untouched by the CML API.
	drift := labHasDrift(&lab, desired)
	tflog.Info(ctx, "Resource LabLifecycle UPDATE: sync decision", map[string]any{
		"desired":       desired,
		"state_changed": stateChanged,
		"drift":         drift,
		"lab_state":     lab.State,
		"nodes":         len(lab.Nodes),
		"links":         len(lab.Links),
	})

	if stateChanged || desired == models.LabStateStarted || drift {
		tflog.Info(
			ctx, "Resource LabLifecycle UPDATE: applying state change or correcting drift",
			map[string]any{"desired": desired, "state_changed": stateChanged, "wait": wait, "lab_state": lab.State},
		)

		start := startData{
			lab:      &lab,
			staging:  getStaging(ctx, req.Config, &resp.Diagnostics),
			timeouts: getTimeouts(ctx, req.Config, &resp.Diagnostics),
			wait:     wait,
		}

		reconcileLinks := func(current *models.Lab, want models.LabState) {
			for _, l := range current.Links {
				switch want {
				case models.LabStateStarted:
					if l.State != models.LinkStateStarted {
						linkErr := r.cfg.Client().Link.Start(ctx, current.ID, l.ID)
						if linkErr != nil {
							resp.Diagnostics.AddError(
								common.ErrorLabel,
								fmt.Sprintf("Unable to start link %s, got error: %s", l.ID, linkErr),
							)
						}
					}
				case models.LabStateStopped:
					if l.State != models.LinkStateStopped {
						linkErr := r.cfg.Client().Link.Stop(ctx, current.ID, l.ID)
						if linkErr != nil {
							resp.Diagnostics.AddError(
								common.ErrorLabel,
								fmt.Sprintf("Unable to stop link %s, got error: %s", l.ID, linkErr),
							)
						}
					}
				}
			}
		}

		switch desired {
		case models.LabStateStarted:
			r.startNodes(ctx, &resp.Diagnostics, start)
			// Explicitly reconcile links: lab/node start is not sufficient when a
			// link was manually stopped out-of-band.
			reconcileLinks(&lab, desired)
			if start.wait {
				timeout := start.timeouts.Update.ValueString()
				common.Converge(ctx, r.cfg.Client(), &resp.Diagnostics, planData.LabID.ValueString(), timeout)
			}
		case models.LabStateStopped:
			r.stop(ctx, resp.Diagnostics, planData.LabID.ValueString())
			reconcileLinks(&lab, desired)
			if start.wait {
				timeout := start.timeouts.Update.ValueString()
				common.Converge(ctx, r.cfg.Client(), &resp.Diagnostics, planData.LabID.ValueString(), timeout)
			}
		case models.LabStateDefined:
			// Wipe requires a stop first if the lab (or any node) is still running.
			if lab.State == models.LabStateStarted || lab.Running() {
				r.stop(ctx, resp.Diagnostics, planData.LabID.ValueString())
				if start.wait {
					timeout := start.timeouts.Update.ValueString()
					common.Converge(ctx, r.cfg.Client(), &resp.Diagnostics, planData.LabID.ValueString(), timeout)
				}
			}
			r.wipe(ctx, resp.Diagnostics, planData.LabID.ValueString())
			if start.wait {
				timeout := start.timeouts.Update.ValueString()
				common.Converge(ctx, r.cfg.Client(), &resp.Diagnostics, planData.LabID.ValueString(), timeout)
			}
		}

		// Re-read after action so we reflect the post-apply state.
		lab, err = r.cfg.Client().Lab.GetByID(ctx, models.UUID(planData.LabID.ValueString()), true)
		if err != nil {
			resp.Diagnostics.AddError(
				common.ErrorLabel,
				fmt.Sprintf("Unable to fetch lab after action, got error: %s", err),
			)
			return
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Update: lab state: %s", lab.State))

	// If the user explicitly configured a desired state, keep it in state after
	// apply to avoid "inconsistent result" errors when the simulator returns a
	// transitional/lagging state (e.g. wait=false).
	if configData.State.IsNull() {
		planData.State = types.StringValue(string(lab.State))
	}
	planData.Nodes = r.populateNodes(ctx, &lab, &resp.Diagnostics)
	planData.Booted = types.BoolValue(lab.Booted())

	resp.Diagnostics.Append(resp.State.Set(ctx, planData)...)
	tflog.Info(ctx, "Resource LabLifecycle UPDATE done")
}
