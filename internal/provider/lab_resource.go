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
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

const CML2ErrorLabel = "CML2 Provider Error"

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &LabResource{}
var _ resource.ResourceWithImportState = &LabResource{}
var _ tfsdk.AttributeValidator = labStateValidator{}

type LabResource struct {
	client *cmlclient.Client
}

func NewLabResource() resource.Resource {
	return &LabResource{}
}

func (r *LabResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lab"
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

		booted, err = r.client.ConvergedLab(ctx, id)
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
	err := r.client.StopLab(ctx, id)
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
	err := r.client.WipeLab(ctx, id)
	if err != nil {
		diags.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab wipe done")
}

func (r *LabResource) start(ctx context.Context, diags diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab start")
	err := r.client.StartLab(ctx, id)
	if err != nil {
		diags.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to start lab, got error: %s", err),
		)
	}
	tflog.Info(ctx, "lab start done")
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
		err = r.client.SetNodeConfig(ctx, node, config_string)
		if err != nil {
			diags.AddError("set node config failed",
				fmt.Sprintf("setting the new node configuration failed: %s", err),
			)
		}
	}
	tflog.Info(ctx, "injectConfigs: done")
}

func (r *LabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *LabResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Create: import")
	lab, err := r.client.ImportLab(ctx, data.Topology.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to import lab, got error: %s", err),
		)
		return
	}

	// if unspecified, start the lab...
	r.injectConfigs(ctx, lab, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.State.Null || data.State.Value == cmlclient.LabStateStarted {
		r.start(ctx, resp.Diagnostics, lab.ID)
	}

	// if unspecified, wait for it to converge
	if data.Wait.Null || data.Wait.Value {
		timeouts := getTimeouts(ctx, req.Config, &resp.Diagnostics)
		r.converge(ctx, &resp.Diagnostics, lab.ID, timeouts.Create)
	}

	// fetch lab again, with nodes and interfaces
	lab, err = r.client.GetLab(ctx, lab.ID, false)
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

	lab, err := r.client.GetLab(ctx, data.Id.Value, false)
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
	var configData, planData, stateData *LabResourceModel

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

		timeouts := getTimeouts(ctx, req.Config, &resp.Diagnostics)

		// this is very blunt ...
		if stateData.State.Value == cmlclient.LabStateStarted {
			if planData.State.Value == cmlclient.LabStateStopped {
				r.stop(ctx, resp.Diagnostics, planData.Id.Value)
			}
			if planData.State.Value == cmlclient.LabStateDefined {
				r.stop(ctx, resp.Diagnostics, planData.Id.Value)
				r.converge(ctx, &resp.Diagnostics, planData.Id.Value, timeouts.Update)
				r.wipe(ctx, resp.Diagnostics, planData.Id.Value)
			}
		}

		if stateData.State.Value == cmlclient.LabStateStopped {
			if planData.State.Value == cmlclient.LabStateStarted {
				r.start(ctx, resp.Diagnostics, planData.Id.Value)
			}
			if planData.State.Value == cmlclient.LabStateDefined {
				r.wipe(ctx, resp.Diagnostics, planData.Id.Value)
			}
		}

		if stateData.State.Value == cmlclient.LabStateDefined {
			if planData.State.Value == cmlclient.LabStateStarted {
				r.start(ctx, resp.Diagnostics, planData.Id.Value)
			}
		}
		// not sure if this makes sense... state could change when not waiting
		// for convergence.  then again, there's no differentiation at the lab
		// level between "STARTED" and "BOOTED" (e.g. converged).  It's always
		// started...
		if planData.Wait.Null || planData.Wait.Value {
			r.converge(ctx, &resp.Diagnostics, planData.Id.Value, timeouts.Update)
		}
	}

	// since we have changed lab state, we need to re-read all the node
	// state...
	lab, err := r.client.GetLab(ctx, planData.Id.Value, false)
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

	lab, err := r.client.GetLab(ctx, data.Id.Value, true)
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

	err = r.client.DestroyLab(ctx, data.Id.Value)
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
