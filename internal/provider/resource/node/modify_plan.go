package node

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

// ModifyPlan is called when the provider has an opportunity to modify
// the plan: once during the plan phase when Terraform is determining
// the diff that should be shown to the user for approval, and once
// during the apply phase with any unknown values from configuration
// filled in with their final values.
//
// The planned new state is represented by
// ModifyPlanResponse.Plan. It must meet the following
// constraints:
// 1. Any non-Computed attribute set in config must preserve the exact
// config value or return the corresponding attribute value from the
// prior state (ModifyPlanRequest.State).
// 2. Any attribute with a known value must not have its value changed
// in subsequent calls to ModifyPlan or Create/Read/Update.
// 3. Any attribute with an unknown value may either remain unknown
// or take on any value of the expected type.
//
// Any errors will prevent further resource-level plan modifications.

func (r *NodeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	// var stateData, planData cmlschema.NodeModel
	var planData cmlschema.NodeModel

	tflog.Info(ctx, "Resource Node MODIFYPLAN")

	// when deleting, there's no plan
	if req.Plan.Raw.IsNull() {
		tflog.Info(ctx, "there is no plan")
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
	if planData.ImageDefinition.IsUnknown() {
		planData.ImageDefinition = types.StringNull()
	}
	if planData.SerialDevices.IsUnknown() {
		planData.SerialDevices = types.ListNull(cmlschema.SerialDevicesAttrType)
	}
	if planData.RAM.IsUnknown() {
		planData.RAM = types.Int64Null()
	}
	if planData.CPUs.IsUnknown() {
		planData.CPUs = types.Int64Null()
	}
	if planData.VNCkey.IsUnknown() {
		planData.VNCkey = types.StringNull()
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Node MODIFYPLAN done")
}
