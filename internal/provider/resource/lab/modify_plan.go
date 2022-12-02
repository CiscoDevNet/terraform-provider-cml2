package lab

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

func (r *LabResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	var stateData, planData cmlschema.LabModel

	tflog.Info(ctx, "Resource Lab MODIFYPLAN")

	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform current state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if state and plan are identical -> modified date has changed
	// this gets auto-updated when we change something
	if !reflect.DeepEqual(stateData, planData) {
		planData.Modified = types.StringUnknown()
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, planData)...)
	tflog.Info(ctx, "Resource Lab MODIFYPLAN: done")
}
