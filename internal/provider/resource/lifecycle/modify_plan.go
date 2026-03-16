package lifecycle

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *LabLifecycleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var configData, planData, stateData cmlschema.LabLifecycleModel

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
	if planData.State.ValueString() == string(models.LabStateStopped) {
		if !noState && stateData.State.ValueString() == string(models.LabStateDefined) {
			resp.Diagnostics.AddError(
				common.ErrorLabel,
				"can't transition from DEFINED_ON_CORE to STOPPED",
			)
			return
		}
		if noState && planData.State.ValueString() == string(models.LabStateStopped) {
			resp.Diagnostics.AddError(
				common.ErrorLabel,
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

		var nodes map[string]cmlschema.NodeModel

		resp.Diagnostics.Append(tfsdk.ValueAs(ctx, stateData.Nodes, &nodes)...)
		if resp.Diagnostics.HasError() {
			return
		}

		plannedState := planData.State.ValueString()

		for id, node := range nodes {
			// Coordinates can change outside of Terraform (manual drag/drop in UI or
			// auto-layout). During a lifecycle state transition, avoid pinning x/y to
			// prior known values or Terraform may report an "inconsistent result after
			// apply" when the controller returns updated coordinates.
			node.X = types.Int64Unknown()
			node.Y = types.Int64Unknown()

			if plannedState == string(models.LabStateDefined) {
				node.SerialDevices = types.ListNull(cmlschema.SerialDevicesAttrType)
				node.VNCkey = types.StringNull()
				node.ComputeID = types.StringNull()
				node.DataVolume = types.Int64Null()
				node.CPUs = types.Int64Null()
				node.RAM = types.Int64Null()
				node.BootDiskSize = types.Int64Null()
				node.State = types.StringValue(string(models.NodeStateDefined))
				// The controller commonly returns cpu_limit=100 even when a node is in
				// DEFINED_ON_CORE. Keep the planned value aligned with what we will
				// observe after apply to avoid "inconsistent result after apply".
				if !common.IsBuiltInNodeDefinition(node.NodeDefinition.ValueString()) {
					// Determine node type from node definition metadata.
					var libvirtDriver, linuxDriver string
					var isLibvirt bool
					var ndFound bool
					var fetchErr error
					if defs, err := r.cfg.NodeDefinitions(ctx); err == nil {
						if nd, ok := defs[models.UUID(node.NodeDefinition.ValueString())]; ok {
							ndFound = true
							libvirtDriver = common.NodeDefLibvirtDomainDriver(nd)
							linuxDriver = common.NodeDefLinuxDriver(nd)
							isLibvirt = common.NodeDefIsLibvirtBacked(nd)
						}
					} else {
						fetchErr = err
					}
					tflog.Info(ctx, "lifecycle node type probe (cpu_limit heuristic)", map[string]any{
						"lab_state":             plannedState,
						"node_id":               id,
						"nodedefinition":        node.NodeDefinition.ValueString(),
						"nodedefinition_found":  ndFound,
						"nodedefinition_error":  common.ErrorString(fetchErr),
						"libvirt_domain_driver": libvirtDriver,
						"driver":                linuxDriver,
						"is_libvirt":            isLibvirt,
					})
					if isLibvirt {
						node.CPUlimit = types.Int64Value(100)
					} else {
						node.CPUlimit = types.Int64Null()
					}
				} else {
					node.CPUlimit = types.Int64Null()
				}
				node.ImageDefinition = types.StringNull()
			}
			if plannedState == string(models.LabStateStarted) {
				node.SerialDevices = types.ListUnknown(cmlschema.SerialDevicesAttrType)
				node.VNCkey = types.StringUnknown()
				node.ComputeID = types.StringUnknown()
				node.DataVolume = types.Int64Unknown()
				node.CPUs = types.Int64Unknown()
				node.RAM = types.Int64Unknown()
				node.BootDiskSize = types.Int64Unknown()
				node.State = types.StringUnknown()
				// The controller returns cpu_limit=100 for most started nodes even if the
				// field was null in the prior state. If we keep it unknown here, the
				// nested schema's UseStateForUnknown modifier will pin it to null and
				// Terraform will error with "inconsistent result after apply".
				if !common.IsBuiltInNodeDefinition(node.NodeDefinition.ValueString()) {
					// Determine node type from node definition metadata.
					var libvirtDriver, linuxDriver string
					var isLibvirt bool
					if defs, err := r.cfg.NodeDefinitions(ctx); err == nil {
						if nd, ok := defs[models.UUID(node.NodeDefinition.ValueString())]; ok {
							libvirtDriver = common.NodeDefLibvirtDomainDriver(nd)
							linuxDriver = common.NodeDefLinuxDriver(nd)
							isLibvirt = common.NodeDefIsLibvirtBacked(nd)
						}
					}
					tflog.Debug(ctx, "lifecycle node cpu_limit heuristic", map[string]any{
						"lab_state":             plannedState,
						"node_id":               id,
						"nodedefinition":        node.NodeDefinition.ValueString(),
						"libvirt_domain_driver": libvirtDriver,
						"driver":                linuxDriver,
						"is_libvirt":            isLibvirt,
					})
					if isLibvirt {
						node.CPUlimit = types.Int64Value(100)
					} else {
						node.CPUlimit = types.Int64Null()
					}
				} else {
					node.CPUlimit = types.Int64Null()
				}
				node.ImageDefinition = types.StringUnknown()
			}
			if plannedState == string(models.LabStateStopped) {
				if node.State.ValueString() != string(models.NodeStateDefined) {
					node.State = types.StringValue(string(models.NodeStateStopped))
				}
			}

			// This is a bit of a hack since the node def name is hard coded
			// here.  What happens is that UMS nodes get the bridge name as the
			// configuration.  So, we start with no configuration and after
			// start, the configuration is set to the name of the bridge, like
			// ums-b843d547-54.
			// As an alternative, all configurations could be set to "Unknown"
			if node.NodeDefinition.ValueString() == "unmanaged_switch" {
				node.Configuration = cmlschema.NewConfigUnknown()
				node.Configurations = types.ListNull(cmlschema.NamedConfigAttrType)
			}

			var ifaces []cmlschema.InterfaceModel
			resp.Diagnostics.Append(tfsdk.ValueAs(ctx, node.Interfaces, &ifaces)...)
			if resp.Diagnostics.HasError() {
				return
			}

			for idx := range ifaces {
				if plannedState == string(models.LabStateStarted) {
					ifaces[idx].IP4 = types.ListUnknown(types.StringType)
					ifaces[idx].IP6 = types.ListUnknown(types.StringType)
					// MACaddresses won't change at state change if one was assigned
					if ifaces[idx].MACaddress.IsNull() {
						ifaces[idx].MACaddress = types.StringUnknown()
					}
					ifaces[idx].State = types.StringUnknown()
				}
				if plannedState == string(models.LabStateDefined) || plannedState == string(models.LabStateStopped) {
					ifaces[idx].IP4 = types.ListNull(types.StringType)
					ifaces[idx].IP6 = types.ListNull(types.StringType)
				}
				if plannedState == string(models.LabStateDefined) {
					ifaces[idx].MACaddress = types.StringNull()
					ifaces[idx].State = types.StringValue(string(models.IfaceStateDefined))
				}
				if plannedState == string(models.LabStateStopped) {
					if ifaces[idx].State.ValueString() != string(models.IfaceStateDefined) {
						ifaces[idx].State = types.StringValue(string(models.IfaceStateStopped))
					}
				}
			}

			resp.Diagnostics.Append(
				tfsdk.ValueFrom(
					ctx,
					ifaces,
					types.ListType{ElemType: types.ObjectType{AttrTypes: cmlschema.InterfaceAttrType}},
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
				types.MapType{ElemType: types.ObjectType{AttrTypes: cmlschema.NodeAttrType}},
				&planData.Nodes,
			)...,
		)
		if resp.Diagnostics.HasError() {
			return
		}

		// booted state of lab is unknown if the plan is to start
		if planData.State.ValueString() == string(models.LabStateStarted) {
			planData.Booted = types.BoolUnknown()
		} else {
			planData.Booted = types.BoolValue(false)
		}
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Lifecycle MODIFYPLAN done")
}
