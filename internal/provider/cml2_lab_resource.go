package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = cmlLabResourceType{}
var _ tfsdk.Resource = cmlLabResource{}
var _ tfsdk.ResourceWithImportState = cmlLabResource{}
var _ tfsdk.AttributeValidator = labStateValidator{}

type cmlLabResourceType struct{}

type labStateValidator struct{}

const CML2ErrorLabel = "CML2 Provider Error"

func (v labStateValidator) Description(ctx context.Context) string {
	return "valid states are DEFINED_ON_CORE, STOPPED and STARTED"
}

// MarkdownDescription returns a markdown formatted description of the
// validator's behavior, suitable for a practitioner to understand its impact.
func (v labStateValidator) MarkdownDescription(ctx context.Context) string {
	return "valid states are `DEFINED_ON_CORE`, `STOPPED` and `STARTED`"
}

// Validate runs the main validation logic of the validator, reading
// configuration data out of `req` and updating `resp` with diagnostics.
func (v labStateValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	tflog.Info(ctx, "##### i am here")
	var labState types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &labState)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	tflog.Info(ctx, "##### now here")

	if labState.Unknown || labState.Null {
		return
	}

	tflog.Info(ctx, "##### finally here:["+labState.Value+"]")

	if labState.Value != cmlclient.LabStateDefined &&
		labState.Value != cmlclient.LabStateStopped &&
		labState.Value != cmlclient.LabStateStarted {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid lab state",
			"valid states are DEFINED_ON_CORE, STOPPED and STARTED.",
		)
		return
	}
}

func (t cmlLabResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CML Lab resource",

		Attributes: map[string]tfsdk.Attribute{
			// topology is mostly marked as sensitive b/c lengthy topo
			// yaml clutters the output
			"topology": {
				MarkdownDescription: "topology to start",
				Required:            true,
				Type:                types.StringType,
				Sensitive:           true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"start": {
				MarkdownDescription: "topology will be started if true",
				Optional:            true,
				Type:                types.BoolType,
			},
			"wait": {
				MarkdownDescription: "wait until topology is BOOTED if true",
				Optional:            true,
				Type:                types.BoolType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "CML lab identifier, a UUID",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"state": {
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "CML lab state",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
				Validators: []tfsdk.AttributeValidator{
					labStateValidator{},
				},
			},
		},
	}, nil
}

func (t cmlLabResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return cmlLabResource{
		provider: provider,
	}, diags
}

type cmlLabResourceData struct {
	Topology types.String `tfsdk:"topology"`
	Start    types.Bool   `tfsdk:"start"`
	Wait     types.Bool   `tfsdk:"wait"`
	Id       types.String `tfsdk:"id"`
	State    types.String `tfsdk:"state"`
}

type cmlLabResource struct {
	provider provider
}

func (r cmlLabResource) converge(ctx context.Context, diag diag.Diagnostics, id string) {
	converged := false
	waited := 0
	snoozeFor := 5 // seconds
	var err error

	tflog.Info(ctx, "waiting for convergence")

	for !converged {

		converged, err = r.provider.client.ConvergedLab(id)
		if err != nil {
			diag.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Wait for convergence of lab, got error: %s", err),
			)
			return
		}
		time.Sleep(time.Second * time.Duration(snoozeFor))
		waited++
		tflog.Info(
			ctx, "converging",
			map[string]interface{}{"seconds": waited * snoozeFor},
		)
	}
}

func (r cmlLabResource) ModifyPlan(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {

	var (
		configData cmlLabResourceData
		planData   cmlLabResourceData
		stateData  cmlLabResourceData
	)

	tflog.Info(ctx, "modify plan")

	diags := req.Config.Get(ctx, &configData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// do we have state?
	noState := req.State.Raw.IsNull()
	if !noState {
		diags = req.State.Get(ctx, &stateData)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// check if a specified configuration is valid
	if noState && !configData.State.Null {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			"Can't set state when lab isn't yet created!",
		)
		return
	}

	// get the planned state
	diags = resp.Plan.Get(ctx, &planData)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// check if we can transition to specified state
	if !noState && planData.State.Value == cmlclient.LabStateStopped {
		if stateData.State.Value == cmlclient.LabStateDefined {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				"can't transition from DEFINED_ON_CORE to STOPPED",
			)
			return
		}
	}

	// is a change of the start attribute planned?
	if !noState && planData.Start.Value != stateData.Start.Value && !stateData.Id.Null {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			"The resource is already created, it doesn't make sense to change the start attribute.",
		)
		return
	}

	tflog.Info(ctx, "modify plan done")
}

func (r cmlLabResource) stop(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab stop")
	err := r.provider.client.StopLab(id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to stop CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab stop done")
}

func (r cmlLabResource) wipe(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab wipe")
	err := r.provider.client.WipeLab(id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab wipe done")
}

func (r cmlLabResource) start(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab start")
	err := r.provider.client.StartLab(id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to start lab, got error: %s", err),
		)
	}
	tflog.Info(ctx, "lab start done")
}

func (r cmlLabResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data cmlLabResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "lab import")
	lab, err := r.provider.client.ImportLab(data.Topology.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to import lab, got error: %s", err),
		)
		return
	}

	if data.Start.Null || data.Start.Value {
		r.start(ctx, resp.Diagnostics, lab.ID)
	}

	if data.Wait.Null || data.Wait.Value {
		r.converge(ctx, resp.Diagnostics, lab.ID)
	}

	// fetch lab again
	lab, err = r.provider.client.GetLab(lab.ID, true)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	data.Id = types.String{Value: lab.ID}
	data.State = types.String{Value: lab.State}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	tflog.Info(ctx, "lab create done")
}

func (r cmlLabResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data cmlLabResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "lab read")

	lab, err := r.provider.client.GetLab(data.Id.Value, true)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to read CML2 lab, got error: %s", err),
		)
		return
	}
	data.Id = types.String{Value: lab.ID}
	data.State = types.String{Value: lab.State}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	tflog.Info(ctx, "lab read done")
}

func (r cmlLabResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data, current cmlLabResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &current)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if current.State.Value != data.State.Value {
		tflog.Info(ctx, "state changed")

		// this is very blunt ...

		if current.State.Value == cmlclient.LabStateStarted {
			if data.State.Value == cmlclient.LabStateStopped {
				r.stop(ctx, resp.Diagnostics, data.Id.Value)
			}
			if data.State.Value == cmlclient.LabStateDefined {
				r.stop(ctx, resp.Diagnostics, data.Id.Value)
				r.converge(ctx, resp.Diagnostics, data.Id.Value)
				r.wipe(ctx, resp.Diagnostics, data.Id.Value)
			}
		}

		if current.State.Value == cmlclient.LabStateStopped {
			if data.State.Value == cmlclient.LabStateStarted {
				r.start(ctx, resp.Diagnostics, data.Id.Value)
			}
			if data.State.Value == cmlclient.LabStateDefined {
				r.wipe(ctx, resp.Diagnostics, data.Id.Value)
			}
		}

		if current.State.Value == cmlclient.LabStateDefined {
			if data.State.Value == cmlclient.LabStateStarted {
				r.start(ctx, resp.Diagnostics, data.Id.Value)
			}
		}
		// not sure if this makes sense... state could change when not waiting
		// for convergence.  then again, there's no differentiation at the lab
		// level between "STARTED" and "BOOTED" (e.g. converged).  It's always
		// started...
		if data.Wait.Null || data.Wait.Value {
			r.converge(ctx, resp.Diagnostics, data.Id.Value)
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	tflog.Info(ctx, "update a resource")
}

func (r cmlLabResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data cmlLabResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := r.provider.client.GetLab(data.Id.Value, true)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to read CML2 lab, got error: %s", err),
		)
		return
	}

	if lab.State != cmlclient.LabStateDefined {
		if lab.State == cmlclient.LabStateStarted {
			r.stop(ctx, resp.Diagnostics, data.Id.Value)
		}
		r.wipe(ctx, resp.Diagnostics, data.Id.Value)
	}

	err = r.provider.client.DestroyLab(data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab resource destroyed")
}

func (r cmlLabResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
