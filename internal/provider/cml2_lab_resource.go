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

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = cml2LabResourceType{}
var _ tfsdk.Resource = cmlLabResource{}
var _ tfsdk.ResourceWithImportState = cmlLabResource{}
var _ tfsdk.AttributeValidator = labStateValidator{}

type cml2LabResourceType struct{}

type labStateValidator struct{}

const CML2ErrorLabel = "CML2 Provider Error"

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
					tfsdk.ListNestedAttributesOptions{},
				),
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (t cml2LabResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return cmlLabResource{
		provider: provider,
	}, diags
}

type cmlLabResourceData struct {
	Topology types.String `tfsdk:"topology"`
	Wait     types.Bool   `tfsdk:"wait"`
	Id       types.String `tfsdk:"id"`
	State    types.String `tfsdk:"state"`
	Nodes    types.List   `tfsdk:"nodes"`
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

	// if no TF state and there should be a lab state set
	// if noState && !configData.State.Null {
	// 	resp.Diagnostics.AddError(
	// 		CML2ErrorLabel,
	// 		"Can't set lab state when it isn't yet created!",
	// 	)
	// 	return
	// }

	// get the planned state
	diags = resp.Plan.Get(ctx, &planData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "ModifyPlan: plan has errors")
		return
	}

	// check if we can transition to specified state
	if !noState && planData.State.Value == cmlclient.LabStateStopped {
		if stateData.State.Value == cmlclient.LabStateDefined {
			resp.Diagnostics.AddError(
				CML2ErrorLabel,
				"can't transition from DEFINED_ON_CORE to STOPPED",
			)
			return
		}
	}

	if !noState && planData.State.Value != stateData.State.Value {
		tflog.Info(ctx, "ModifyPlan: state change")

		// this doesn't work as I'm not changing the actually data :(
		for _, nodeElem := range planData.Nodes.Elems {
			node := resultNode{}
			nodeElem.(types.Object).As(ctx, node, types.ObjectAsOptions{})
			node.State.Unknown = true
			for _, ifaceElem := range node.Interfaces.Elems {
				iface := resultInterface{}
				ifaceElem.(types.Object).As(ctx, iface, types.ObjectAsOptions{})
				iface.State.Unknown = true
				iface.MACaddress.Unknown = true
				iface.IP4 = nil
			}
		}
		// planData.Nodes.Unknown = true
		diags = resp.Plan.Set(ctx, planData)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "ModifyPlan: plan has errors")
			return
		}
	}

	// tflog.Info(ctx, "ModifyPlan: done", map[string]interface{}{
	// 	"nodes": planData.Nodes,
	// })
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
	data.Nodes.Elems = populateNodes(ctx, lab)
	data.Nodes.Null = false

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	tflog.Info(ctx, "Create: done")
}

var (
	ifaceObject = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":           types.StringType,
			"label":        types.StringType,
			"state":        types.StringType,
			"mac_address":  types.StringType,
			"is_connected": types.BoolType,
			"ip4": types.ListType{
				ElemType: types.StringType,
			},
			"ip6": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
	nodeObject = types.Object{
		AttrTypes: map[string]attr.Type{
			"id":       types.StringType,
			"label":    types.StringType,
			"state":    types.StringType,
			"nodetype": types.StringType,
			"interfaces": types.ListType{
				ElemType: ifaceObject,
			},
		},
	}
)

func populateNodes(ctx context.Context, lab *cmlclient.Lab) []attr.Value {
	// we want this as a stable sort by node UUID
	nodeList := []*cmlclient.Node{}
	for _, node := range lab.Nodes {
		nodeList = append(nodeList, node)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})

	nodes := make([]attr.Value, 0)
	for _, node := range nodeList {

		// we want this as a stable sort by interface UUID
		ilist := []*cmlclient.Interface{}
		for _, iface := range node.Interfaces {
			ilist = append(ilist, iface)
		}
		sort.Slice(ilist, func(i, j int) bool {
			return ilist[i].ID < ilist[j].ID
		})

		ifaces := make([]attr.Value, 0)
		for _, iface := range ilist {

			ip4list := make([]attr.Value, 0)
			for _, ip := range iface.IP4 {
				ip4list = append(ip4list, types.String{Value: ip})
			}
			ip6list := make([]attr.Value, 0)
			for _, ip := range iface.IP6 {
				ip6list = append(ip6list, types.String{Value: ip})
			}

			ifaceElem := types.Object{
				AttrTypes: ifaceObject.AttrTypes,
				Attrs: map[string]attr.Value{
					"id":           types.String{Value: iface.ID},
					"label":        types.String{Value: iface.Label},
					"state":        types.String{Value: iface.State},
					"mac_address":  types.String{Value: iface.MACaddress},
					"is_connected": types.Bool{Value: iface.IsConnected},
					"ip4": types.List{
						ElemType: types.StringType,
						Elems:    ip4list,
						Null:     false,
					},
					"ip6": types.List{
						ElemType: types.StringType,
						Elems:    ip6list,
						Null:     false,
					},
				},
			}
			ifaces = append(ifaces, ifaceElem)
		}

		o := types.Object{
			AttrTypes: nodeObject.AttrTypes,
			Attrs: map[string]attr.Value{
				"id":       types.String{Value: node.ID},
				"label":    types.String{Value: node.Label},
				"state":    types.String{Value: node.State},
				"nodetype": types.String{Value: node.NodeDefinition},
				"interfaces": types.List{
					ElemType: ifaceObject,
					Elems:    ifaces,
					Null:     false,
				},
			},
		}
		// tflog.Info(ctx, "node add", map[string]interface{}{
		// 	"object": o,
		// })
		nodes = append(nodes, o)
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
	data.Nodes.Elems = populateNodes(ctx, lab)
	data.Nodes.Null = false

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Read: errors!")
		return
	}
	tflog.Info(ctx, "Read: done")
}

func (r cmlLabResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data, current cmlLabResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &current)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if current.State.Value != data.State.Value {
		tflog.Info(ctx, "state changed")

		// this is very blunt ...
		if current.State.Value == cmlclient.LabStateStarted {
			if data.State.Value == cmlclient.LabStateStopped {
				r.stop(ctx, resp.Diagnostics, data.Id.Value)
			}
			if data.State.Value == cmlclient.LabStateDefined {
				r.stop(ctx, resp.Diagnostics, data.Id.Value)
				r.converge(ctx, resp.Diagnostics, data.Id.Value)
				r.wipe(ctx, resp.Diagnostics, data.Id.Value)
			}
		}

		if current.State.Value == cmlclient.LabStateStopped {
			if data.State.Value == cmlclient.LabStateStarted {
				r.start(ctx, resp.Diagnostics, data.Id.Value)
			}
			if data.State.Value == cmlclient.LabStateDefined {
				r.wipe(ctx, resp.Diagnostics, data.Id.Value)
			}
		}

		if current.State.Value == cmlclient.LabStateDefined {
			if data.State.Value == cmlclient.LabStateStarted {
				r.start(ctx, resp.Diagnostics, data.Id.Value)
			}
		}
		// not sure if this makes sense... state could change when not waiting
		// for convergence.  then again, there's no differentiation at the lab
		// level between "STARTED" and "BOOTED" (e.g. converged).  It's always
		// started...
		if data.Wait.Null || data.Wait.Value {
			r.converge(ctx, resp.Diagnostics, data.Id.Value)
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	tflog.Info(ctx, "update a resource")
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
