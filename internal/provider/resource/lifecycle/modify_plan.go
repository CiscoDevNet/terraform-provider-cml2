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

	var configData, planData, stateData *schema.LabLifecycleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "ModifyPlan")

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
	resp.Diagnostics.Append(resp.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "ModifyPlan: plan has errors")
		return
	}

	if planData == nil {
		tflog.Error(ctx, "ModifyPlan: no plan exists...")
		return
	}

	// if !noState && !configData.Elements.IsNull() {
	// 	for idx := 0; idx < len(planData.Elements.Elems); idx++ {
	// 		id := stateData.Elements.Elems[idx].(types.String)
	// 		id.Null = false
	// 		id.Unknown = false
	// 		resp.Diagnostics.Append(tfsdk.ValueFrom(ctx, id, types.StringType, &planData.Elements.Elems[idx])...)
	// 		if resp.Diagnostics.HasError() {
	// 			return
	// 		}
	// 	}
	// }

	// check if we can transition to specified state
	if planData.State.Value == cmlclient.LabStateStopped {
		if !noState && stateData.State.Value == cmlclient.LabStateDefined {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				"can't transition from DEFINED_ON_CORE to STOPPED",
			)
			return
		}
		if noState && planData.State.Value == cmlclient.LabStateStopped {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				"can't transition from no state to STOPPED",
			)
			return
		}
	}

	changeNeeded := false
	if !noState {
		changeNeeded = planData.State.Value != stateData.State.Value
	}

	if changeNeeded {
		tflog.Info(ctx, "ModifyPlan: change detected")

		var nodes map[string]schema.NodeModel

		resp.Diagnostics.Append(tfsdk.ValueAs(ctx, planData.Nodes, &nodes)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for id, node := range nodes {

			// these all need to be re-read when state changes...  based on
			// actual state change, these can be optimized to provide a better
			// state diff -- but it works for now
			node.State.Unknown = true
			node.State.Null = false

			node.DataVolume.Unknown = true
			node.DataVolume.Null = false

			node.ComputeID.Unknown = true
			node.ComputeID.Null = false

			node.SerialDevices.Unknown = true
			node.SerialDevices.Null = false

			node.CPUs.Unknown = true
			node.CPUs.Null = false

			node.VNCkey.Unknown = true
			node.VNCkey.Null = false

			node.RAM.Unknown = true
			node.RAM.Null = false

			node.BootDiskSize.Unknown = true
			node.BootDiskSize.Null = false

			// This is a bit of a hack since the node def name is hard coded
			// here.  what happens is that UMS nodes get the bridge name as the
			// configuration.  So, we start with no configuration and after
			// start, the configuration is set to the name of the bridge, like
			// ums-b843d547-54.
			// As an alternative, all configurations could be set to "Unknown"
			if node.NodeDefinition.Value == "unmanaged_switch" {
				node.Configuration.Unknown = true
			}

			var ifaces []schema.InterfaceModel
			resp.Diagnostics.Append(tfsdk.ValueAs(ctx, node.Interfaces, &ifaces)...)
			if resp.Diagnostics.HasError() {
				return
			}

			for idx := range ifaces {
				ifaces[idx].IP4.Unknown = true
				ifaces[idx].IP6.Unknown = true
				// we know that when we wipe, the MAC is going to be null
				if planData.State.Value == "DEFINED_ON_CORE" {
					ifaces[idx].MACaddress.Unknown = false
					ifaces[idx].MACaddress.Null = true
				} else {
					// MACaddresses won't change at state change if one was assigned
					if ifaces[idx].MACaddress.Null {
						ifaces[idx].MACaddress.Unknown = true
						ifaces[idx].MACaddress.Null = false // why? oh, why!
					}
				}
				ifaces[idx].State.Unknown = true

				// iface := ifaces[idx]
				// tflog.Info(ctx, fmt.Sprintf("mac: %v/%v", iface.MACaddress.Null, iface.MACaddress.Unknown))
				// tflog.Info(ctx, fmt.Sprintf("ip4: %v/%v", iface.IP4.Null, iface.IP4.Unknown))
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

		// booted state of lab is unknown at this point
		planData.Booted.Unknown = true
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, planData)...)
	tflog.Info(ctx, "ModifyPlan: done")
}
