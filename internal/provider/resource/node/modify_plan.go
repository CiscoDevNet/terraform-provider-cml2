package node

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r *NodeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	var stateData, planData *schema.NodeModel

	tflog.Info(ctx, "Resource Node ModifyPlan")

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

	if planData.DataVolume.Unknown {
		planData.DataVolume.Unknown = false
		planData.DataVolume.Null = true
	}
	if planData.BootDiskSize.Unknown {
		planData.BootDiskSize.Unknown = false
		planData.BootDiskSize.Null = true
	}
	if planData.ComputeID.Unknown {
		planData.ComputeID.Unknown = false
		planData.ComputeID.Null = true
	}

	// tflog.Info(ctx, fmt.Sprintf("XXX Plan: %+v", planData.DataVolume))
	// tflog.Info(ctx, fmt.Sprintf("XXX State: %+v", stateData.DataVolume))

	resp.Diagnostics.Append(resp.Plan.Set(ctx, planData)...)

	tflog.Info(ctx, "Resource Node ModifyPlan: done")
}
