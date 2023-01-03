package link

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

func (r *LinkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	var planData, stateData cmlschema.LinkModel

	tflog.Info(ctx, "Resource Link MODIFYPLAN")

	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// these are the fields which are optional and computed... if they are
	// specified, then we need to copy over the state data into the plan

	if !stateData.SlotA.IsUnknown() {
		planData.SlotA = stateData.SlotA
	}

	if !stateData.SlotB.IsUnknown() {
		planData.SlotB = stateData.SlotB
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &planData)...)
	tflog.Info(ctx, "Resource Link MODIFYPLAN done")
}
