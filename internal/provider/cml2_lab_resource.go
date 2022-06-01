package provider

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

const CML2ErrorLabel = "CML2 Provider Error"

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = cml2LabResourceType{}
var _ tfsdk.Resource = cmlLabResource{}
var _ tfsdk.ResourceWithImportState = cmlLabResource{}
var _ tfsdk.AttributeValidator = labStateValidator{}

type cml2LabResourceType struct{}

type labStateValidator struct{}

func (v labStateValidator) Description(ctx context.Context) string {
	return "valid states are DEFINED_ON_CORE, STOPPED and STARTED"
}

// MarkdownDescription returns a markdown formatted description of the
// validator's behavior, suitable for a practitioner to understand its impact.
func (v labStateValidator) MarkdownDescription(ctx context.Context) string {
	return "valid states are `DEFINED_ON_CORE`, `STOPPED` and `STARTED`"
}

// Validate runs the main validation logic of the validator, reading
// configuration data out of `req` and updating `resp` with diagnostics.
func (v labStateValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var labState types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &labState)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if labState.Unknown || labState.Null {
		return
	}

	if labState.Value != cmlclient.LabStateDefined &&
		labState.Value != cmlclient.LabStateStopped &&
		labState.Value != cmlclient.LabStateStarted {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid lab state",
			"valid states are DEFINED_ON_CORE, STOPPED and STARTED.",
		)
		return
	}
}

func (t cml2LabResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {

	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CML Lab resource",

		Attributes: map[string]tfsdk.Attribute{
			// topology is marked as sensitive mostly b/c lengthy topology
			// YAML clutters the output.
			"topology": {
				MarkdownDescription: "topology to start",
				Required:            true,
				Type:                types.StringType,
				Sensitive:           true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"wait": {
				MarkdownDescription: "wait until topology is BOOTED if true",
				Optional:            true,
				Type:                types.BoolType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "CML lab identifier, a UUID",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"converged": {
				Computed:            true,
				MarkdownDescription: "CML lab has converged (e.g. BOOTED)",
				Type:                types.BoolType,
				// PlanModifiers: tfsdk.AttributePlanModifiers{
				// 	tfsdk.UseStateForUnknown(),
				// },
			},
			"state": {
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "CML lab state",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
				Validators: []tfsdk.AttributeValidator{
					labStateValidator{},
				},
			},
			"nodes": {
				MarkdownDescription: "List of nodes and their interfaces with IP addresses",
				Computed:            true,
				Attributes: tfsdk.ListNestedAttributes(
					nodeSchema(),
				),
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"configurations": {
				MarkdownDescription: "List of node configurations to store into nodes",
				Optional:            true,
				Type: types.MapType{
					ElemType: types.StringType,
				},
			},
		},
	}, nil
}

func interfaceSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			MarkdownDescription: "Interface ID (UUID)",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"mac_address": {
			MarkdownDescription: "MAC address",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"is_connected": {
			MarkdownDescription: "connection status",
			Type:                types.BoolType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"ip4": {
			MarkdownDescription: "IPv4 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"ip6": {
			MarkdownDescription: "IPv6 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
	}
}

func nodeSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			MarkdownDescription: "Node ID (UUID)",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"configuration": {
			MarkdownDescription: "configuration",
			Type:                types.StringType,
			Computed:            true,
			Sensitive:           true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"nodetype": {
			MarkdownDescription: "Node Type / Definition",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
		"interfaces": {
			MarkdownDescription: "interfaces on the node",
			Computed:            true,
			Attributes: tfsdk.ListNestedAttributes(
				interfaceSchema(),
			),
			PlanModifiers: tfsdk.AttributePlanModifiers{
				tfsdk.UseStateForUnknown(),
			},
		},
	}
}

func (t cml2LabResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return cmlLabResource{
		provider: provider,
	}, diags
}

type cmlLabResourceData struct {
	Topology       types.String `tfsdk:"topology"`
	Wait           types.Bool   `tfsdk:"wait"`
	Id             types.String `tfsdk:"id"`
	State          types.String `tfsdk:"state"`
	Nodes          types.List   `tfsdk:"nodes"`
	Converged      types.Bool   `tfsdk:"converged"`
	Configurations types.Map    `tfsdk:"configurations"`
}

type cml2Node struct {
	Id            types.String `tfsdk:"id"`
	Label         types.String `tfsdk:"label"`
	State         types.String `tfsdk:"state"`
	NodeType      types.String `tfsdk:"nodetype"`
	Interfaces    types.List   `tfsdk:"interfaces"`
	Configuration types.String `tfsdk:"configuration"`
}

var (
	ifaceObject = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":           types.StringType,
			"label":        types.StringType,
			"state":        types.StringType,
			"mac_address":  types.StringType,
			"is_connected": types.BoolType,
			"ip4":          types.ListType{ElemType: types.StringType},
			"ip6":          types.ListType{ElemType: types.StringType},
		},
	}
	nodeObject = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":            types.StringType,
			"label":         types.StringType,
			"state":         types.StringType,
			"nodetype":      types.StringType,
			"configuration": types.StringType,
			"interfaces":    types.ListType{ElemType: ifaceObject},
		},
	}
)

func newCML2node(node *cmlclient.Node) cml2Node {

	// we want this as a stable sort by interface UUID
	ilist := []*cmlclient.Interface{}
	for _, iface := range node.Interfaces {
		ilist = append(ilist, iface)
	}
	sort.Slice(ilist, func(i, j int) bool {
		return ilist[i].ID < ilist[j].ID
	})

	ifaces := types.List{ElemType: ifaceObject}
	for _, iface := range ilist {
		ifaces.Elems = append(ifaces.Elems, newCML2iface(iface).toObject())
	}

	return cml2Node{
		Id:            types.String{Value: node.ID},
		Label:         types.String{Value: node.Label},
		State:         types.String{Value: node.State},
		NodeType:      types.String{Value: node.NodeDefinition},
		Configuration: types.String{Value: node.Configuration},
		Interfaces:    ifaces,
	}
}

func (n cml2Node) toObject() types.Object {
	return types.Object{
		AttrTypes: nodeObject.AttrTypes,
		Attrs: map[string]attr.Value{
			"id":            n.Id,
			"label":         n.Label,
			"state":         n.State,
			"nodetype":      n.NodeType,
			"configuration": n.Configuration,
			"interfaces":    n.Interfaces,
		},
	}
}

type cml2Interface struct {
	Id          types.String `tfsdk:"id"`
	Label       types.String `tfsdk:"label"`
	State       types.String `tfsdk:"state"`
	MACaddress  types.String `tfsdk:"mac_address"`
	IsConnected types.Bool   `tfsdk:"is_connected"`
	IP4         types.List   `tfsdk:"ip4"`
	IP6         types.List   `tfsdk:"ip6"`
}

func newCML2iface(iface *cmlclient.Interface) cml2Interface {

	ip4List := types.List{ElemType: types.StringType, Null: true}
	ip6List := types.List{ElemType: types.StringType, Null: true}
	macAddress := types.String{Null: true}

	if iface.Runs() {
		// IPv4 addresses
		list := make([]attr.Value, 0)
		for _, ip := range iface.IP4 {
			list = append(list, types.String{Value: ip})
		}
		ip4List.Elems = list
		ip4List.Null = false
		// IPv6 addresses
		list = make([]attr.Value, 0)
		for _, ip := range iface.IP6 {
			list = append(list, types.String{Value: ip})
		}
		ip6List.Elems = list
		ip6List.Null = false
	}
	if iface.Exists() {
		macAddress.Value = iface.MACaddress
		macAddress.Null = false
	}

	return cml2Interface{
		Id:          types.String{Value: iface.ID},
		Label:       types.String{Value: iface.Label},
		State:       types.String{Value: iface.State},
		IsConnected: types.Bool{Value: iface.IsConnected},
		MACaddress:  macAddress,
		IP4:         ip4List,
		IP6:         ip6List,
	}
}

func (n cml2Interface) toObject() types.Object {
	return types.Object{
		AttrTypes: ifaceObject.AttrTypes,
		Attrs: map[string]attr.Value{
			"id":           n.Id,
			"label":        n.Label,
			"state":        n.State,
			"is_connected": n.IsConnected,
			"mac_address":  n.MACaddress,
			"ip4":          n.IP4,
			"ip6":          n.IP6,
		},
	}
}

type cmlLabResource struct {
	provider cml2
}

func (r cmlLabResource) converge(ctx context.Context, diag diag.Diagnostics, id string) {
	converged := false
	waited := 0
	snoozeFor := 5 // seconds
	var err error

	tflog.Info(ctx, "waiting for convergence")

	for !converged {

		converged, err = r.provider.client.ConvergedLab(ctx, id)
		if err != nil {
			diag.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Wait for convergence of lab, got error: %s", err),
			)
			return
		}
		time.Sleep(time.Second * time.Duration(snoozeFor))
		waited++
		tflog.Info(
			ctx, "converging",
			map[string]interface{}{"seconds": waited * snoozeFor},
		)
	}
}

func (r cmlLabResource) ModifyPlan(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {

	var (
		configData cmlLabResourceData
		planData   cmlLabResourceData
		stateData  cmlLabResourceData
	)

	tflog.Info(ctx, "ModifyPlan")

	diags := req.Config.Get(ctx, &configData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// do we have state?
	noState := req.State.Raw.IsNull()
	if !noState {
		diags = req.State.Get(ctx, &stateData)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// get the planned state
	diags = resp.Plan.Get(ctx, &planData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "ModifyPlan: plan has errors")
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

	if !noState && planData.State.Value != stateData.State.Value {
		tflog.Info(ctx, "ModifyPlan: state change")

		newNodeList := types.List{ElemType: nodeObject}
		nodes := []cml2Node{}
		diags := planData.Nodes.ElementsAs(ctx, &nodes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "ModifyPlan: that didn't work")
			return
		}

		for _, node := range nodes {

			// check if configs should be changed
			// if
			newInterfaceList := types.List{ElemType: ifaceObject}
			for _, ifaceElem := range node.Interfaces.Elems {

				iface := cml2Interface{}
				diags = ifaceElem.(types.Object).As(ctx, &iface, types.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					tflog.Error(ctx, "ModifyPlan: that didn't work")
					return
				}

				macAddress := types.String{}
				ip4List := types.List{ElemType: types.StringType}
				ip6List := types.List{ElemType: types.StringType}

				switch planData.State.Value {
				case cmlclient.LabStateDefined:
					macAddress.Null = true
					ip4List.Null = true
					ip6List.Null = true
					if config, ok := configData.Configurations.Elems[node.Label.Value]; ok {
						var configuration string
						diags := tfsdk.ValueAs(ctx, config, &configuration)
						resp.Diagnostics.Append(diags...)
						if resp.Diagnostics.HasError() {
							return
						}
						node.Configuration.Value = configuration
					}
				case cmlclient.LabStateStarted:
					if stateData.State.Value == cmlclient.LabStateDefined {
						macAddress.Unknown = true
					} else {
						macAddress.Value = iface.MACaddress.Value
					}
					ip4List.Unknown = true
					ip6List.Unknown = true
				case cmlclient.LabStateStopped:
					macAddress.Value = iface.MACaddress.Value
					ip4List.Null = true
					ip6List.Null = true
				}
				iface.State.Unknown = true
				iface.MACaddress = macAddress
				iface.IP4 = ip4List
				iface.IP6 = ip6List
				newIfaceElem := iface.toObject()
				newInterfaceList.Elems = append(newInterfaceList.Elems, newIfaceElem)
			}

			node.State.Unknown = true
			node.Interfaces = newInterfaceList
			newNodeElem := node.toObject()
			newNodeList.Elems = append(newNodeList.Elems, newNodeElem)
		}

		// modify node state
		ap := tftypes.NewAttributePath().WithAttributeName("nodes")
		diags = resp.Plan.SetAttribute(ctx, ap, newNodeList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "ModifyPlan: nodes plan has errors")
			return
		}
		// modify converged state
		ap = tftypes.NewAttributePath().WithAttributeName("converged")
		diags = resp.Plan.SetAttribute(ctx, ap, types.Bool{Unknown: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "ModifyPlan: converged plan has errors")
			return
		}
	}
	tflog.Info(ctx, "ModifyPlan: done")
}

func (r cmlLabResource) stop(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab stop")
	err := r.provider.client.StopLab(ctx, id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to stop CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab stop done")
}

func (r cmlLabResource) wipe(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab wipe")
	err := r.provider.client.WipeLab(ctx, id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab wipe done")
}

func (r cmlLabResource) start(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab start")
	err := r.provider.client.StartLab(ctx, id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to start lab, got error: %s", err),
		)
	}
	tflog.Info(ctx, "lab start done")
}

func (r cmlLabResource) injectConfigs(ctx context.Context, lab *cmlclient.Lab, data cmlLabResourceData) diag.Diagnostics {
	tflog.Info(ctx, "injectConfigs")

	// nodes := []cml2Node{}
	// diags := data.Nodes.ElementsAs(ctx, &nodes, false)
	// if diags.HasError() {
	// 	return diags
	// }

	// for _, node := range nodes {
	// 	tflog.Info(ctx, fmt.Sprintf("node: %+v", node))
	// 	if !node.Configuration.Null {
	// 		if actualNode, ok := lab.Nodes[node.Id.Value]; ok {
	// 			if actualNode.Configuration != node.Configuration.Value {
	// 				err := r.provider.client.SetNodeConfig(ctx, lab.ID, node.Id.Value, node.Configuration.Value)
	// 				if err != nil {
	// 					diags.AddError("set node config failed",
	// 						fmt.Sprintf("setting the new node configuration for %s failed: %s", node.Label.Value, err),
	// 					)
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	var diags diag.Diagnostics
	for _, node := range lab.Nodes {
		// set the configuration IF the node is not yet started
		tflog.Info(ctx, fmt.Sprintf("injectConfigs: node state, %+v", node.State))
		if node.State == cmlclient.NodeStateDefined {
			tflog.Info(ctx, fmt.Sprintf("injectConfigs: defined, %+v", node))
			if config, ok := data.Configurations.Elems[node.Label]; ok {
				var configuration string
				diags := tfsdk.ValueAs(ctx, config, &configuration)
				diags.Append(diags...)
				if diags.HasError() {
					return diags
				}
				tflog.Info(ctx, fmt.Sprintf("injectConfigs: %s, %s", configuration, node.Configuration))
				if configuration == node.Configuration {
					continue
				}
				err := r.provider.client.SetNodeConfig(ctx, lab.ID, node.ID, configuration)
				if err != nil {
					diags.AddError("set node config failed",
						fmt.Sprintf("setting the new node configuration failed: %s", err),
					)
				}
			}
		}
	}
	tflog.Info(ctx, "injectConfigs: done")

	return diags
}

func (r cmlLabResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data cmlLabResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Create: import")
	lab, err := r.provider.client.ImportLab(ctx, data.Topology.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to import lab, got error: %s", err),
		)
		return
	}

	r.injectConfigs(ctx, lab, data)
	if data.State.Null || data.State.Value == cmlclient.LabStateStarted {
		r.start(ctx, resp.Diagnostics, lab.ID)
	}

	if data.Wait.Null || data.Wait.Value {
		r.converge(ctx, resp.Diagnostics, lab.ID)
	}

	// fetch lab again, with nodes and interfaces
	lab, err = r.provider.client.GetLab(ctx, lab.ID, false)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to get lab, got error: %s", err),
		)
		return
	}

	data.Id = types.String{Value: lab.ID}
	data.State = types.String{Value: lab.State}
	data.Nodes = populateNodes(ctx, lab)
	data.Converged = types.Bool{Value: lab.Booted()}

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	tflog.Info(ctx, "Create: done")
}

func populateNodes(ctx context.Context, lab *cmlclient.Lab) types.List {
	// we want this as a stable sort by node UUID
	nodeList := []*cmlclient.Node{}
	for _, node := range lab.Nodes {
		nodeList = append(nodeList, node)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})
	nodes := types.List{ElemType: nodeObject}
	for _, node := range nodeList {
		o := newCML2node(node).toObject()
		nodes.Elems = append(nodes.Elems, o)
	}
	return nodes
}

func (r cmlLabResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data cmlLabResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}

	tflog.Info(ctx, "Read: start")

	lab, err := r.provider.client.GetLab(ctx, data.Id.Value, false)
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
	data.Nodes = populateNodes(ctx, lab)
	data.Converged = types.Bool{Value: lab.Booted()}

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}
	tflog.Info(ctx, "Read: done")
}

func (r cmlLabResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var planData, stateData cmlLabResourceData

	diags := req.Plan.Get(ctx, &planData)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &stateData)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(planData.Configurations.Elems) > 0 {
		lab, err := r.provider.client.GetLab(ctx, planData.Id.Value, false)
		if err != nil {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to fetch lab, got error: %s", err),
			)
			return
		}
		r.injectConfigs(ctx, lab, planData)
	}

	if stateData.State.Value != planData.State.Value {
		tflog.Info(ctx, "state changed")

		// this is very blunt ...
		if stateData.State.Value == cmlclient.LabStateStarted {
			if planData.State.Value == cmlclient.LabStateStopped {
				r.stop(ctx, resp.Diagnostics, planData.Id.Value)
			}
			if planData.State.Value == cmlclient.LabStateDefined {
				r.stop(ctx, resp.Diagnostics, planData.Id.Value)
				r.converge(ctx, resp.Diagnostics, planData.Id.Value)
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
			r.converge(ctx, resp.Diagnostics, planData.Id.Value)
		}

		// since we have changed lab state, we need to re-read all the node
		// state...
		lab, err := r.provider.client.GetLab(ctx, planData.Id.Value, false)
		if err != nil {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				fmt.Sprintf("Unable to fetch lab, got error: %s", err),
			)
			return
		}
		tflog.Info(ctx, fmt.Sprintf("Update: lab state: %s", lab.State))
		planData.State = types.String{Value: lab.State}
		planData.Nodes = populateNodes(ctx, lab)
		planData.Converged = types.Bool{Value: lab.Booted()}
	}

	diags = resp.State.Set(ctx, planData)
	resp.Diagnostics.Append(diags...)
	tflog.Info(ctx, "Update: done")
}

func (r cmlLabResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data cmlLabResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	lab, err := r.provider.client.GetLab(ctx, data.Id.Value, true)
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

	err = r.provider.client.DestroyLab(ctx, data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab resource destroyed")
}

func (r cmlLabResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
