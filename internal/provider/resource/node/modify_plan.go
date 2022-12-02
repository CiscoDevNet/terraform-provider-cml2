package node

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

func (r *NodeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	var planData cmlschema.NodeModel

	tflog.Info(ctx, "Resource Node MODIFYPLAN")

	// when deleting, there's no plan
	if req.Plan.Raw.IsNull() {
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planData.DataVolume.IsUnknown() {
		planData.DataVolume = types.Int64Null()
	}
	if planData.BootDiskSize.IsUnknown() {
		planData.BootDiskSize = types.Int64Null()
	}
	if planData.ComputeID.IsUnknown() {
		planData.ComputeID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Node MODIFYPLAN: done")
}
