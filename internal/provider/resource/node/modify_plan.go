package node

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
)

// ModifyPlan is called when the provider has an opportunity to modify the
// plan: once during the plan phase when Terraform is determining the diff that
// should be shown to the user for approval, and once during the apply phase
// with any unknown values from configuration filled in with their final
// values.
//
// The planned new state is represented by ModifyPlanResponse.Plan. It must
// meet the following constraints:
//
// 1. Any non-Computed attribute set in config must preserve the exact config
// value or return the corresponding attribute value from the prior state
// (ModifyPlanRequest.State).
// 2. Any attribute with a known value must not have its value changed in
// subsequent calls to ModifyPlan or Create/Read/Update.
// 3. Any attribute with an unknown value may either remain unknown or take on
// any value of the expected type.
//
// Any errors will prevent further resource-level plan modifications.

func (r *NodeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// var stateData, planData cmlschema.NodeModel
	var configData, planData, stateData cmlschema.NodeModel

	tflog.Info(ctx, "Resource Node MODIFYPLAN")

	// when deleting, there's no plan
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		tflog.Info(ctx, "there is no plan or state")
		return
	}

	// Read Terraform config/plan/state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// if the node was started once (e.g. not in state DEFINED anymore) then
	// changing certain attributes require a replace
	nodeExists := stateData.State.ValueString() != cmlclient.NodeStateDefined

	// tflog.Warn(ctx, "### CONDITION ###", map[string]any{
	// 	"type_state": stateData.NodeDefinition.ValueString(),
	// 	"type_plan": planData.NodeDefinition.ValueString(),
	// 	"config_state": stateData.Configuration.ValueString(),
	// 	"config_plan": planData.Configuration.ValueString(),
	// 	"unknown_plan": planData.Configuration.IsUnknown(),
	// })

	if nodeExists && !stateData.Configuration.Equal(planData.Configuration) {
		// tflog.Info(ctx, "$$$1")
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("configuration"))
	}

	if nodeExists && !stateData.Configurations.Equal(planData.Configurations) {
		// tflog.Info(ctx, "$$$2")
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("configurations"))
	}

	if !configData.ImageDefinition.IsNull() && !configData.ImageDefinition.Equal(stateData.ImageDefinition) {
		if nodeExists {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("imagedefinition"))
		}
		planData.ImageDefinition = configData.ImageDefinition
	}
	if planData.ImageDefinition.IsUnknown() {
		planData.ImageDefinition = types.StringNull()
	}

	if !configData.RAM.IsNull() && !configData.RAM.Equal(stateData.RAM) {
		if nodeExists {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("ram"))
		}
		planData.RAM = configData.RAM
	}
	if planData.RAM.IsUnknown() {
		planData.RAM = types.Int64Null()
	}

	if !configData.CPUs.IsNull() && !configData.CPUs.Equal(stateData.CPUs) {
		if nodeExists {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("cpus"))
		}
		planData.CPUs = configData.CPUs
	}
	if planData.CPUs.IsUnknown() {
		planData.CPUs = types.Int64Null()
	}

	if !configData.CPUlimit.IsNull() && !configData.CPUlimit.Equal(stateData.CPUlimit) {
		if nodeExists {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("cpu_limit"))
		}
		planData.CPUlimit = configData.CPUlimit
	}
	if planData.CPUlimit.IsUnknown() && planData.NodeDefinition.ValueString() != "external_connector" && planData.NodeDefinition.ValueString() != "unmanaged_switch" {
		// CPUlimit is the weird one here as it is possible to set the value to null
		// and this actually works on updating on the controller (PATCH). However,
		// when reading the data again, the value comes back as 100.
		// See SIMPLE-5052 and cmlclient.NodeGet()
		// Also: Need to restrict this to devices other than UMS and ExtConn as those
		// do always return NULL for the value w/ 2.6.0
		// TODO: need to see what IOL returns in 2.7.0
		planData.CPUlimit = types.Int64Value(int64(100))
	}

	if !configData.DataVolume.IsNull() && !configData.DataVolume.Equal(stateData.DataVolume) {
		if nodeExists {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("data_volume"))
		}
		planData.DataVolume = configData.DataVolume
	}
	if planData.DataVolume.IsUnknown() {
		planData.DataVolume = types.Int64Null()
	}

	if !configData.BootDiskSize.IsNull() && !configData.BootDiskSize.Equal(stateData.BootDiskSize) {
		if nodeExists {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("boot_disk_size"))
		}
		planData.BootDiskSize = configData.BootDiskSize
	}
	if planData.BootDiskSize.IsUnknown() {
		planData.BootDiskSize = types.Int64Null()
	}

	// need to set an empty set if no configuration is provided for tags
	// this is a one-off since tags can't be null
	if configData.Tags.IsNull() {
		tags, dia := types.SetValueFrom(ctx, types.StringType, []string{})
		planData.Tags = tags
		resp.Diagnostics.Append(dia...)
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Node MODIFYPLAN done")
}
