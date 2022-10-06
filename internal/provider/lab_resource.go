package provider

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/m/v2/pkg/cmlclient"
)

const CML2ErrorLabel = "CML2 Provider Error"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &LabResource{}
var _ resource.ResourceWithImportState = &LabResource{}
var _ resource.ResourceWithValidateConfig = &LabResource{}
var _ tfsdk.AttributeValidator = labStateValidator{}

type LabResource struct {
	client *cmlclient.Client
}

type startData struct {
	wait     bool
	lab      *cmlclient.Lab
	staging  *ResourceStaging
	timeouts *ResourceTimeouts
}

func NewLabResource() resource.Resource {
	return &LabResource{}
}

func (r *LabResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lab"
}

func (r *LabResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data LabResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If staging is not configured, return without warning.
	// (I think it never can be unknown as it's configuration data)
	if data.Staging.IsNull() || data.Staging.IsUnknown() {
		return
	}

	// If wait is set (true), return without warning
	// if it is null, then the default is "true" (e.g. wait)
	if data.Wait.IsNull() || data.Wait.Value {
		return
	}

	resp.Diagnostics.AddAttributeWarning(
		path.Root("staging"),
		"Conflicting configuration",
		"Expected \"wait\" to be true with when staging is configured. "+
			"The resource may return unexpected results.",
	)
}

func (r *LabResource) converge(ctx context.Context, diags *diag.Diagnostics, id string, timeout types.String) {
	booted := false
	waited := 0
	snoozeFor := 5 // seconds
	var err error

	tflog.Info(ctx, "waiting for convergence")

	tov, err := time.ParseDuration(timeout.Value)
	if err != nil {
		panic("can't parse timeout -- should be validated")
	}
	endTime := time.Now().Add(tov)

	for !booted {

		booted, err = r.client.HasLabConverged(ctx, id)
		if err != nil {
			diags.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Wait for convergence of lab, got error: %s", err),
			)
			return
		}

		select {
		case <-time.After(time.Second * time.Duration(snoozeFor)):
		case <-ctx.Done():
			return
		}
		if time.Now().After(endTime) {
			diags.AddError(CML2ErrorLabel, fmt.Sprintf("ran into timeout (max %s)", timeout.Value))
			return
		}
		waited++
		tflog.Info(
			ctx, "converging",
			map[string]interface{}{"seconds": waited * snoozeFor},
		)
	}
}

func (r *LabResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	var configData, planData, stateData *LabResourceModel

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

		var nodes map[string]NodeResourceModel

		resp.Diagnostics.Append(tfsdk.ValueAs(ctx, planData.Nodes, &nodes)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for id, node := range nodes {
			node.State.Unknown = true

			// This is a bit of a hack since the node def name is hard coded
			// here.  what happens is that UMS nodes get the bridge name as the
			// configuration.  So, we start with no configuration and after
			// start, the configuration is set to the name of the bridge, like
			// ums-b843d547-54.
			// As an alternative, all configurations could be set to "Unknown"
			if node.NodeDefinition.Value == "unmanaged_switch" {
				node.Configuration.Unknown = true
			}

			var ifaces []InterfaceResourceModel
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
					types.ListType{ElemType: types.ObjectType{AttrTypes: interfaceAttrType}},
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
				types.MapType{ElemType: types.ObjectType{AttrTypes: nodeAttrType}},
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

func (r *LabResource) stop(ctx context.Context, diags diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab stop")
	err := r.client.LabStop(ctx, id)
	if err != nil {
		diags.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to stop CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab stop done")
}

func (r *LabResource) wipe(ctx context.Context, diags diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab wipe")
	err := r.client.LabWipe(ctx, id)
	if err != nil {
		diags.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab wipe done")
}

func (r *LabResource) startNodesAll(ctx context.Context, diags *diag.Diagnostics, start startData) {
	tflog.Info(ctx, "lab start")
	err := r.client.LabStart(ctx, start.lab.ID)
	if err != nil {
		diags.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to start lab, got error: %s", err),
		)
	}
	tflog.Info(ctx, "lab start done")
	if start.wait {
		r.converge(ctx, diags, start.lab.ID, start.timeouts.Create)
	}
}

func (r *LabResource) startNodes(ctx context.Context, diags *diag.Diagnostics, start startData) {

	// start all nodes at once, no staging
	if start.staging == nil {
		r.startNodesAll(ctx, diags, start)
		return
	}

	// start nodes in stages
	for _, stage_elem := range start.staging.Stages.Elems {
		stage := stage_elem.(types.String).Value
		for _, node := range start.lab.Nodes {
			for _, tag := range node.Tags {
				if tag == stage {
					tflog.Info(ctx, fmt.Sprintf("starting node %s", node.Label))
					err := r.client.NodeStart(ctx, node)
					if err != nil {
						diags.AddError(
							CML2ErrorLabel,
							fmt.Sprintf("Unable to start node %s, got error: %s", node.Label, err),
						)
					}
				}
			}
		}
		// this is not 100% correct as the timeout is applied to each stage
		// should be: timeout applied to all stages combined
		r.converge(ctx, diags, start.lab.ID, start.timeouts.Create)
	}

	// start remaining nodes, if indicated
	if start.staging.StartRemaining.Value {
		tflog.Info(ctx, "starting remaining nodes")
		r.startNodesAll(ctx, diags, start)
	}
}

func (r *LabResource) injectConfigs(ctx context.Context, lab *cmlclient.Lab, data *LabResourceModel, diags *diag.Diagnostics) {
	tflog.Info(ctx, "injectConfigs")

	if data.Configs.IsNull() {
		tflog.Info(ctx, "injectConfigs: no configs")
		return
	}

	for nodeID, config := range data.Configs.Elems {
		node, err := lab.NodeByLabel(ctx, nodeID)
		if err == cmlclient.ErrElementNotFound {
			node = lab.Nodes[nodeID]
		}
		if node == nil {
			diags.AddError(CML2ErrorLabel, fmt.Sprintf("node with label %s not found", nodeID))
			continue
		}
		if node.State != cmlclient.NodeStateDefined {
			diags.AddError(CML2ErrorLabel, fmt.Sprintf("unexpected node state %s", node.State))
			continue
		}
		config_string := config.(types.String).Value
		err = r.client.NodeSetConfig(ctx, node, config_string)
		if err != nil {
			diags.AddError("set node config failed",
				fmt.Sprintf("setting the new node configuration failed: %s", err),
			)
		}
	}
	tflog.Info(ctx, "injectConfigs: done")
}

func (r *LabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		data *LabResourceModel
		err  error
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	start := startData{
		staging:  getStaging(ctx, req.Config, &resp.Diagnostics),
		timeouts: getTimeouts(ctx, req.Config, &resp.Diagnostics),
		wait:     data.Wait.Null || data.Wait.Value,
	}

	tflog.Info(ctx, "Create: import")
	start.lab, err = r.client.LabImport(ctx, data.Topology.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to import lab, got error: %s", err),
		)
		return
	}

	// inject the configurations into the nodes
	r.injectConfigs(ctx, start.lab, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// if unknown state or specifically "start" state, start the lab...
	if data.State.Unknown || data.State.Value == cmlclient.LabStateStarted {
		r.startNodes(ctx, &resp.Diagnostics, start)
	}

	// fetch lab again, with nodes and interfaces
	lab, err := r.client.LabGet(ctx, start.lab.ID, false)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	data.Id = types.String{Value: lab.ID}
	data.State = types.String{Value: lab.State}
	data.Nodes = r.populateNodes(ctx, lab, &resp.Diagnostics)
	data.Booted = types.Bool{Value: lab.Booted()}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	tflog.Info(ctx, "Create: done")
}

func (r *LabResource) populateNodes(ctx context.Context, lab *cmlclient.Lab, diags *diag.Diagnostics) types.Map {
	// we want this as a stable sort by node UUID
	nodeList := []*cmlclient.Node{}
	for _, node := range lab.Nodes {
		nodeList = append(nodeList, node)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})
	nodes := types.Map{
		ElemType: types.ObjectType{AttrTypes: nodeAttrType},
		Elems:    make(map[string]attr.Value),
	}
	for _, node := range nodeList {
		nodes.Elems[node.ID] = newNode(ctx, node, diags)
	}
	return nodes
}

func (r *LabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *LabResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// tflog.Info(ctx, "state:", map[string]interface{}{"data": data})

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}

	tflog.Info(ctx, "Read: start")

	lab, err := r.client.LabGet(ctx, data.Id.Value, false)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to fetch lab, got error: %s", err),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Read: lab state: %s", lab.State))

	data.Id = types.String{Value: lab.ID}
	data.State = types.String{Value: lab.State}
	data.Nodes = r.populateNodes(ctx, lab, &resp.Diagnostics)
	data.Booted = types.Bool{Value: lab.Booted()}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}
	tflog.Info(ctx, "Read: done")
}

func (r LabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		configData, planData, stateData *LabResourceModel
		err                             error
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if stateData.State.Value != planData.State.Value {
		tflog.Info(ctx, "state changed")

		start := startData{
			staging:  getStaging(ctx, req.Config, &resp.Diagnostics),
			timeouts: getTimeouts(ctx, req.Config, &resp.Diagnostics),
			wait:     planData.Wait.Null || planData.Wait.Value,
		}

		// need to get the lab data here
		start.lab, err = r.client.LabGet(ctx, planData.Id.Value, false)
		if err != nil {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to fetch lab, got error: %s", err),
			)
			return
		}

		// this is very blunt ...
		if stateData.State.Value == cmlclient.LabStateStarted {
			if planData.State.Value == cmlclient.LabStateStopped {
				r.stop(ctx, resp.Diagnostics, planData.Id.Value)
			}
			if planData.State.Value == cmlclient.LabStateDefined {
				r.stop(ctx, resp.Diagnostics, planData.Id.Value)
				r.converge(ctx, &resp.Diagnostics, planData.Id.Value, start.timeouts.Update)
				r.wipe(ctx, resp.Diagnostics, planData.Id.Value)
			}
		}

		if stateData.State.Value == cmlclient.LabStateStopped {
			if planData.State.Value == cmlclient.LabStateStarted {
				r.startNodes(ctx, &resp.Diagnostics, start)
			}
			if planData.State.Value == cmlclient.LabStateDefined {
				r.wipe(ctx, resp.Diagnostics, planData.Id.Value)
			}
		}

		if stateData.State.Value == cmlclient.LabStateDefined {
			if planData.State.Value == cmlclient.LabStateStarted {
				r.startNodes(ctx, &resp.Diagnostics, start)
			}
		}
		// not sure if this makes sense... state could change when not waiting
		// for convergence.  then again, there's no differentiation at the lab
		// level between "STARTED" and "BOOTED" (e.g. converged).  It's always
		// started...
		if start.wait {
			r.converge(ctx, &resp.Diagnostics, planData.Id.Value, start.timeouts.Update)
		}
	}

	// since we have changed lab state, we need to re-read all the node
	// state...
	lab, err := r.client.LabGet(ctx, planData.Id.Value, false)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to fetch lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Update: lab state: %s", lab.State))
	planData.State = types.String{Value: lab.State}
	planData.Nodes = r.populateNodes(ctx, lab, &resp.Diagnostics)
	planData.Booted = types.Bool{Value: lab.Booted()}

	resp.Diagnostics.Append(resp.State.Set(ctx, planData)...)
	tflog.Info(ctx, "Update: done")
}

func (r *LabResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *LabResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := r.client.LabGet(ctx, data.Id.Value, true)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to read CML2 lab, got error: %s", err),
		)
		return
	}

	if lab.State != cmlclient.LabStateDefined {
		if lab.State == cmlclient.LabStateStarted {
			r.stop(ctx, resp.Diagnostics, data.Id.Value)
		}
		r.wipe(ctx, resp.Diagnostics, data.Id.Value)
	}

	err = r.client.LabDestroy(ctx, data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab resource destroyed")
}

func (r LabResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
