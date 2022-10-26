package lab

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r *LabResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	var stateData, planData *schema.LabModel

	tflog.Info(ctx, "Resource Lab ModifyPlan")

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

	// when deleting, there's no plan
	if planData == nil {
		return
	}

	// if state and plan are identical -> modified date has changed
	// this gets auto-updated when we change something
	if !reflect.DeepEqual(stateData, planData) {
		planData.Modified.Unknown = true
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, planData)...)
	tflog.Info(ctx, "Resource Lab ModifyPlan: done")
}
