package lifecycle

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cmlclient "github.com/rschmied/gocmlclient"

	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r *LabLifecycleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	var configData, planData, stateData schema.LabLifecycleModel

	tflog.Info(ctx, "Resource Lifecycle MODIFYPLAN")

	// configuration data for the resource
	if req.Config.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// do we have state?
	noState := req.State.Raw.IsNull()
	if !noState {
		resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// get the planned state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "ModifyPlan: plan has errors")
		return
	}

	// check if we can transition to specified state
	if planData.State.ValueString() == cmlclient.LabStateStopped {
		if !noState && stateData.State.ValueString() == cmlclient.LabStateDefined {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				"can't transition from DEFINED_ON_CORE to STOPPED",
			)
			return
		}
		if noState && planData.State.ValueString() == cmlclient.LabStateStopped {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				"can't transition from no state to STOPPED",
			)
			return
		}
	}

	// store the default in state if it is not provided in the configuration
	if configData.State.IsNull() {
		planData.State = types.StringValue("STARTED")
	}

	changeNeeded := false
	if !noState {
		changeNeeded = planData.State.ValueString() != stateData.State.ValueString()
	}

	if changeNeeded {
		tflog.Info(ctx, "ModifyPlan: change detected")

		var nodes map[string]schema.NodeModel

		resp.Diagnostics.Append(tfsdk.ValueAs(ctx, planData.Nodes, &nodes)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for id, node := range nodes {

			planState := planData.State.ValueString()

			if planData.State.ValueString() == cmlclient.LabStateDefined {
				node.SerialDevices = types.ListNull(schema.SerialDevicesAttrType)
				node.VNCkey = types.StringNull()
				node.ComputeID = types.StringNull()
				node.DataVolume = types.Int64Null()
				node.CPUs = types.Int64Null()
				node.RAM = types.Int64Null()
				node.BootDiskSize = types.Int64Null()
				node.State = types.StringValue(cmlclient.NodeStateDefined)
			}
			if planData.State.ValueString() == cmlclient.LabStateStarted {
				node.SerialDevices = types.ListUnknown(schema.SerialDevicesAttrType)
				node.VNCkey = types.StringUnknown()
				node.ComputeID = types.StringUnknown()
				node.DataVolume = types.Int64Unknown()
				node.CPUs = types.Int64Unknown()
				node.RAM = types.Int64Unknown()
				node.BootDiskSize = types.Int64Unknown()
				node.State = types.StringUnknown()
			}
			if planData.State.ValueString() == cmlclient.LabStateStopped {
				if node.State.ValueString() != cmlclient.NodeStateDefined {
					node.State = types.StringValue(cmlclient.NodeStateStopped)
				}
			}

			// This is a bit of a hack since the node def name is hard coded
			// here.  what happens is that UMS nodes get the bridge name as the
			// configuration.  So, we start with no configuration and after
			// start, the configuration is set to the name of the bridge, like
			// ums-b843d547-54.
			// As an alternative, all configurations could be set to "Unknown"
			if node.NodeDefinition.ValueString() == "unmanaged_switch" {
				node.Configuration = types.StringUnknown()
			}

			var ifaces []schema.InterfaceModel
			resp.Diagnostics.Append(tfsdk.ValueAs(ctx, node.Interfaces, &ifaces)...)
			if resp.Diagnostics.HasError() {
				return
			}

			for idx := range ifaces {
				if planState == cmlclient.LabStateStarted {
					ifaces[idx].IP4 = types.ListUnknown(types.StringType)
					ifaces[idx].IP6 = types.ListUnknown(types.StringType)
					// MACaddresses won't change at state change if one was assigned
					if ifaces[idx].MACaddress.IsNull() {
						ifaces[idx].MACaddress = types.StringUnknown()
					}
					ifaces[idx].State = types.StringUnknown()
				}
				if planState == cmlclient.LabStateDefined || planState == cmlclient.LabStateStopped {
					ifaces[idx].IP4 = types.ListNull(types.StringType)
					ifaces[idx].IP6 = types.ListNull(types.StringType)
				}
				if planState == cmlclient.LabStateDefined {
					ifaces[idx].MACaddress = types.StringNull()
					ifaces[idx].State = types.StringValue(cmlclient.IfaceStateDefined)
				}
				if planState == cmlclient.LabStateStopped {
					if ifaces[idx].State.ValueString() != cmlclient.IfaceStateDefined {
						ifaces[idx].State = types.StringValue(cmlclient.IfaceStateStopped)
					}
				}
			}

			resp.Diagnostics.Append(
				tfsdk.ValueFrom(
					ctx,
					ifaces,
					types.ListType{ElemType: types.ObjectType{AttrTypes: schema.InterfaceAttrType}},
					&node.Interfaces,
				)...,
			)
			if resp.Diagnostics.HasError() {
				return
			}
			nodes[id] = node
		}

		resp.Diagnostics.Append(
			tfsdk.ValueFrom(
				ctx,
				nodes,
				types.MapType{ElemType: types.ObjectType{AttrTypes: schema.NodeAttrType}},
				&planData.Nodes,
			)...,
		)
		if resp.Diagnostics.HasError() {
			return
		}

		// booted state of lab is unknown if the plan is to start
		if planData.State.ValueString() == cmlclient.LabStateStarted {
			planData.Booted = types.BoolUnknown()
		} else {
			planData.Booted = types.BoolValue(false)
		}
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Lifecycle MODIFYPLAN: done")
}
