package provider

import (
	"context"
	"fmt"
	"time"

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

type LabResourceModel struct {
	Topology types.String `tfsdk:"topology"`
	Wait     types.Bool   `tfsdk:"wait"`
	Id       types.String `tfsdk:"id"`
	State    types.String `tfsdk:"state"`
	Booted   types.Bool   `tfsdk:"booted"`
	// Nodes    types.Map    `tfsdk:"nodes"`

	// nodeAttrs      map[string]attr.Type
	// interfaceAttrs map[string]attr.Type

	// Configurations types.Map    `tfsdk:"configurations"`
	// Special types.Map `tfsdk:"special"`
}

func NewLabResource() resource.Resource {
	return &LabResource{}
}

// func NewCML2LabResource() resource.Resource {
// 	interfaceAttr := map[string]attr.Type{
// 		"id":           types.StringType,
// 		"label":        types.StringType,
// 		"state":        types.StringType,
// 		"mac_address":  types.StringType,
// 		"is_connected": types.BoolType,
// 		"ip4":          types.ListType{ElemType: types.StringType},
// 		"ip6":          types.ListType{ElemType: types.StringType},
// 	}
// 	nodeAttr := map[string]attr.Type{
// 		"id":            types.StringType,
// 		"label":         types.StringType,
// 		"state":         types.StringType,
// 		"nodetype":      types.StringType,
// 		"configuration": types.StringType,
// 		"interfaces": types.MapType{
// 			ElemType: types.ObjectType{
// 				AttrTypes: interfaceAttr,
// 			},
// 		},
// 		"tags": types.ListType{ElemType: types.StringType},
// 	}
// 	return LabResource{nodeAttrs: nodeAttr}
// }

func (r *LabResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lab"
}

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

func (t *LabResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {

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
					resource.RequiresReplace(),
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
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"booted": {
				Computed:            true,
				MarkdownDescription: "All nodes in the lab have booted",
				Type:                types.BoolType,
			},
			"state": {
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "CML lab state",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
				Validators: []tfsdk.AttributeValidator{
					labStateValidator{},
				},
			},
			// "nodes": {
			// 	MarkdownDescription: "List of nodes and their interfaces with IP addresses",
			// 	Computed:            true,
			// 	Attributes: tfsdk.MapNestedAttributes(
			// 		nodeSchema(),
			// 	),
			// 	PlanModifiers: tfsdk.AttributePlanModifiers{
			// 		resource.UseStateForUnknown(),
			// 	},
			// },
			// "configurations": {
			// 	MarkdownDescription: "List of node configurations to store into nodes",
			// 	Optional:            true,
			// 	Type: types.MapType{
			// 		ElemType: types.StringType,
			// 	},
			// },
			// "special": {
			// 	MarkdownDescription: "State of specific nodes. The key is either the node name or the name of a tag.  In both cases, a regular expression can be used. If the result is ambiguous, the node name takes preference.",
			// 	Optional:            true,
			// 	// Attributes: tfsdk.MapNestedAttributes(
			// 	// 	specialSchema(),
			// 	// ),
			// 	Type: types.MapType{
			// 		ElemType: types.ObjectType{
			// 			AttrTypes: map[string]attr.Type{
			// 				"configuration": types.StringType,
			// 				"state":         types.StringType,
			// 				"image_id":      types.StringType,
			// 			},
			// 		},
			// 	},
			// },
		},
	}, nil
}

// func specialSchema() map[string]tfsdk.Attribute {
// 	return map[string]tfsdk.Attribute{
// 		"configuration": {
// 			MarkdownDescription: "the configuration of the node",
// 			Type:                types.StringType,
// 			Optional:            true,
// 		},
// 		"state": {
// 			MarkdownDescription: "the desired state of the node",
// 			Type:                types.StringType,
// 			Optional:            true,
// 		},
// 		"image_id": {
// 			MarkdownDescription: "the image_id the node should use",
// 			Type:                types.StringType,
// 			Optional:            true,
// 		},
// 	}
// }

func interfaceSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			MarkdownDescription: "Interface ID (UUID)",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"mac_address": {
			MarkdownDescription: "MAC address",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"is_connected": {
			MarkdownDescription: "connection status",
			Type:                types.BoolType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ip4": {
			MarkdownDescription: "IPv4 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ip6": {
			MarkdownDescription: "IPv6 address list",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
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
				resource.UseStateForUnknown(),
			},
		},
		"label": {
			MarkdownDescription: "label",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"state": {
			MarkdownDescription: "state",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"configuration": {
			MarkdownDescription: "configuration",
			Type:                types.StringType,
			Computed:            true,
			Sensitive:           true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"nodetype": {
			MarkdownDescription: "Node Type / Definition",
			Type:                types.StringType,
			Computed:            true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"interfaces": {
			MarkdownDescription: "interfaces on the node",
			Computed:            true,
			Attributes: tfsdk.ListNestedAttributes(
				interfaceSchema(),
			),
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"tags": {
			MarkdownDescription: "Tags of the node",
			Computed:            true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
	}
}

// type cml2Node struct {
// 	Id            types.String `tfsdk:"id"`
// 	Label         types.String `tfsdk:"label"`
// 	State         types.String `tfsdk:"state"`
// 	NodeType      types.String `tfsdk:"nodetype"`
// 	Tags          types.List   `tfsdk:"tags"`
// 	Interfaces    types.List   `tfsdk:"interfaces"`
// 	Configuration types.String `tfsdk:"configuration"`
// }

// type cml2Special struct {
// 	Configuration types.String `tfsdk:"configuration"`
// 	State         types.String `tfsdk:"state"`
// 	ImageID       types.String `tfsdk:"image_id"`
// }

// type cml2SpecialMap map[string]cml2Special

// var (
// 	ifaceObject = types.ObjectType{
// 		AttrTypes: map[string]attr.Type{
// 			"id":           types.StringType,
// 			"label":        types.StringType,
// 			"state":        types.StringType,
// 			"mac_address":  types.StringType,
// 			"is_connected": types.BoolType,
// 			"ip4":          types.ListType{ElemType: types.StringType},
// 			"ip6":          types.ListType{ElemType: types.StringType},
// 		},
// 	}
// 	nodeObject = types.ObjectType{
// 		AttrTypes: map[string]attr.Type{
// 			"id":            types.StringType,
// 			"label":         types.StringType,
// 			"state":         types.StringType,
// 			"nodetype":      types.StringType,
// 			"configuration": types.StringType,
// 			"interfaces":    types.ListType{ElemType: ifaceObject},
// 			"tags":          types.ListType{ElemType: types.StringType},
// 		},
// 	}
// )

// func (r *LabResource) newCML2node(ctx context.Context, node *cmlclient.Node) cml2Node {

// 	// we want this as a stable sort by interface UUID
// 	ilist := []*cmlclient.Interface{}
// 	for _, iface := range node.Interfaces {
// 		ilist = append(ilist, iface)
// 	}
// 	sort.Slice(ilist, func(i, j int) bool {
// 		return ilist[i].ID < ilist[j].ID
// 	})

// 	ifaces := types.List{ElemType: types.ObjectType{
// 		AttrTypes: r.interfaceAttrs,
// 	}}
// 	for _, iface := range ilist {

// 		newIfaceElem := types.Object{}
// 		diags := tfsdk.ValueFrom(
// 			ctx, iface, types.ObjectType{
// 				AttrTypes: r.interfaceAttrs,
// 			}, &newIfaceElem)
// 		diags.Append(diags...)
// 		if diags.HasError() {
// 			panic("uh-oh")
// 		}

// 		ifaces.Elems = append(ifaces.Elems, newIfaceElem)
// 	}

// 	tags := types.List{ElemType: types.StringType}
// 	for _, tag := range node.Tags {
// 		tags.Elems = append(tags.Elems, types.String{Value: tag})
// 	}

// 	return cml2Node{
// 		Id:            types.String{Value: node.ID},
// 		Label:         types.String{Value: node.Label},
// 		State:         types.String{Value: node.State},
// 		NodeType:      types.String{Value: node.NodeDefinition},
// 		Configuration: types.String{Value: node.Configuration},
// 		Interfaces:    ifaces,
// 		Tags:          tags,
// 	}
// }

type cml2Interface struct {
	Id          types.String `tfsdk:"id"`
	Label       types.String `tfsdk:"label"`
	State       types.String `tfsdk:"state"`
	MACaddress  types.String `tfsdk:"mac_address"`
	IsConnected types.Bool   `tfsdk:"is_connected"`
	IP4         types.List   `tfsdk:"ip4"`
	IP6         types.List   `tfsdk:"ip6"`
}

// func (r *LabResource) newCML2iface(iface *cmlclient.Interface) cml2Interface {

// 	ip4List := types.List{ElemType: types.StringType, Null: true}
// 	ip6List := types.List{ElemType: types.StringType, Null: true}
// 	macAddress := types.String{Null: true}

// 	if iface.Runs() {
// 		// IPv4 addresses
// 		list := make([]attr.Value, 0)
// 		for _, ip := range iface.IP4 {
// 			list = append(list, types.String{Value: ip})
// 		}
// 		ip4List.Elems = list
// 		ip4List.Null = false
// 		// IPv6 addresses
// 		list = make([]attr.Value, 0)
// 		for _, ip := range iface.IP6 {
// 			list = append(list, types.String{Value: ip})
// 		}
// 		ip6List.Elems = list
// 		ip6List.Null = false
// 	}
// 	if iface.Exists() {
// 		macAddress.Value = iface.MACaddress
// 		macAddress.Null = false
// 	}

// 	return cml2Interface{
// 		Id:          types.String{Value: iface.ID},
// 		Label:       types.String{Value: iface.Label},
// 		State:       types.String{Value: iface.State},
// 		IsConnected: types.Bool{Value: iface.IsConnected},
// 		MACaddress:  macAddress,
// 		IP4:         ip4List,
// 		IP6:         ip6List,
// 	}
// }

// func (r *LabResource) matchSpecial(ctx context.Context, diag *diag.Diagnostics, specials cml2SpecialMap, node cml2Node) *cml2Special {
// 	for key, special := range specials {
// 		// check the node label first
// 		matched, err := regexp.Match(key, []byte(node.Label.Value))
// 		if err != nil {
// 			diag.AddError(
// 				CML2ErrorLabel,
// 				fmt.Sprintf("not a valid regex: %s, got %s", key, err),
// 			)
// 			return nil
// 		}
// 		if matched {
// 			return &special
// 		}
// 		// if no match, check all the node tags
// 		for _, tag := range node.Tags.Elems {
// 			matched, err := regexp.Match(key, []byte(tag.(types.String).Value))
// 			if err != nil {
// 				diag.AddError(
// 					CML2ErrorLabel,
// 					fmt.Sprintf("not a valid regex: %s, got %s", key, err),
// 				)
// 				return nil
// 			}
// 			if matched {
// 				return &special
// 			}
// 		}

// 	}
// 	return nil
// }

func (r *LabResource) converge(ctx context.Context, diag diag.Diagnostics, id string) {
	booted := false
	waited := 0
	snoozeFor := 5 // seconds
	var err error

	tflog.Info(ctx, "waiting for convergence")

	for !booted {

		booted, err = r.client.ConvergedLab(ctx, id)
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

	// special := cml2SpecialMap{}
	// if !configData.Special.IsNull() {
	// 	diags = configData.Special.ElementsAs(ctx, &special, false)
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		tflog.Error(ctx, "ModifyPlan: that didn't work")
	// 		return
	// 	}
	// 	tflog.Info(ctx, fmt.Sprintf("SPECIAL: %+v\n", special))
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

	// changeNeeded := false
	// if !noState {
	// 	changeNeeded = planData.State.Value != stateData.State.Value ||
	// 		!planData.Special.Equal(stateData.Special)
	// }

	// if changeNeeded {
	// 	tflog.Info(ctx, "ModifyPlan: change detected")

	// 	newNodeList := types.List{
	// 		ElemType: types.ObjectType{
	// 			AttrTypes: data.nodeAttrs,
	// 		},
	// 	}
	// 	nodes := []cml2Node{}
	// 	diags := planData.Nodes.ElementsAs(ctx, &nodes, false)
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		tflog.Error(ctx, "ModifyPlan: that didn't work")
	// 		return
	// 	}

	// 	for _, node := range nodes {

	// 		sp := r.matchSpecial(ctx, &diags, special, node)
	// 		if diags.HasError() {
	// 			tflog.Error(ctx, "ModifyPlan: specials issue")
	// 			return
	// 		}
	// 		if sp != nil && planData.State.Value == cmlclient.LabStateDefined {
	// 			_ = 1
	// 			if !sp.Configuration.IsNull() {
	// 				node.Configuration.Value = sp.Configuration.Value
	// 			}
	// 			if !sp.State.IsNull() {
	// 				node.State.Value = sp.State.Value
	// 			}
	// 			if !sp.ImageID.IsNull() {
	// 				node.State.Value = sp.ImageID.Value
	// 			}
	// 		}

	// 		newInterfaceList := types.List{
	// 			ElemType: types.ObjectType{
	// 				AttrTypes: data.interfaceAttrs,
	// 			},
	// 		}
	// 		for _, ifaceElem := range node.Interfaces.Elems {

	// 			iface := cml2Interface{}
	// 			diags = ifaceElem.(types.Object).As(ctx, &iface, types.ObjectAsOptions{})
	// 			resp.Diagnostics.Append(diags...)
	// 			if resp.Diagnostics.HasError() {
	// 				tflog.Error(ctx, "ModifyPlan: that didn't work")
	// 				return
	// 			}

	// 			macAddress := types.String{}
	// 			ip4List := types.List{ElemType: types.StringType}
	// 			ip6List := types.List{ElemType: types.StringType}

	// 			switch planData.State.Value {
	// 			case cmlclient.LabStateDefined:
	// 				macAddress.Null = true
	// 				ip4List.Null = true
	// 				ip6List.Null = true
	// 			case cmlclient.LabStateStarted:
	// 				if stateData.State.Value == cmlclient.LabStateDefined {
	// 					macAddress.Unknown = true
	// 				} else {
	// 					macAddress.Value = iface.MACaddress.Value
	// 				}
	// 				ip4List.Unknown = true
	// 				ip6List.Unknown = true
	// 			case cmlclient.LabStateStopped:
	// 				macAddress.Value = iface.MACaddress.Value
	// 				ip4List.Null = true
	// 				ip6List.Null = true
	// 			}
	// 			iface.State.Unknown = true
	// 			iface.MACaddress = macAddress
	// 			iface.IP4 = ip4List
	// 			iface.IP6 = ip6List

	// 			newIfaceElem := types.Object{}
	// 			diags := tfsdk.ValueFrom(
	// 				ctx, node, types.ObjectType{
	// 					AttrTypes: data.interfaceAttrs,
	// 				}, &newIfaceElem)
	// 			resp.Diagnostics.Append(diags...)
	// 			if resp.Diagnostics.HasError() {
	// 				return
	// 			}

	// 			newInterfaceList.Elems = append(newInterfaceList.Elems, newIfaceElem)
	// 		}

	// 		node.Configuration.Unknown = true
	// node.State.Unknown = true
	// 		node.Interfaces = newInterfaceList
	// 		// newNodeElem := node.toObject()

	// 		newNodeElem := types.Object{}
	// 		diags := tfsdk.ValueFrom(
	// 			ctx, node, types.ObjectType{
	// 				AttrTypes: data.nodeAttrs,
	// 			}, &newNodeElem)
	// 		resp.Diagnostics.Append(diags...)
	// 		if resp.Diagnostics.HasError() {
	// 			return
	// 		}

	// 		newNodeList.Elems = append(newNodeList.Elems, newNodeElem)
	// 	}

	// 	// modify node state
	// 	// ap := tftypes.NewAttributePath().WithAttributeName("nodes")
	// 	ap := path.Empty().AtName("root")
	// 	diags = resp.Plan.SetAttribute(ctx, ap, newNodeList)
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		tflog.Error(ctx, "ModifyPlan: nodes plan has errors")
	// 		return
	// 	}
	// 	// modify converged state
	// 	// ap = tftypes.NewAttributePath().WithAttributeName("converged")
	// 	ap = path.Empty().AtName("converged")
	// 	diags = resp.Plan.SetAttribute(ctx, ap, types.Bool{Unknown: true})
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		tflog.Error(ctx, "ModifyPlan: converged plan has errors")
	// 		return
	// 	}
	// }
	tflog.Info(ctx, "ModifyPlan: done")
}

func (r *LabResource) stop(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab stop")
	err := r.client.StopLab(ctx, id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to stop CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab stop done")
}

func (r *LabResource) wipe(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab wipe")
	err := r.client.WipeLab(ctx, id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to destroy CML2 lab, got error: %s", err),
		)
		return
	}
	tflog.Info(ctx, "lab wipe done")
}

func (r *LabResource) start(ctx context.Context, diag diag.Diagnostics, id string) {
	tflog.Info(ctx, "lab start")
	err := r.client.StartLab(ctx, id)
	if err != nil {
		diag.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to start lab, got error: %s", err),
		)
	}
	tflog.Info(ctx, "lab start done")
}

// func (r *LabResource) injectConfigs(ctx context.Context, lab *cmlclient.Lab, data cmlLabResourceData) diag.Diagnostics {
// 	tflog.Info(ctx, "injectConfigs")

// 	var diags diag.Diagnostics

// 	special := cml2SpecialMap{}
// 	if !data.Special.IsNull() {
// 		diags = data.Special.ElementsAs(ctx, &special, false)
// 		diags.Append(diags...)
// 		if diags.HasError() {
// 			tflog.Error(ctx, "injectConfigs: that didn't work (1)")
// 			return diags
// 		}
// 		tflog.Info(ctx, fmt.Sprintf("SPECIAL: %+v\n", special))
// 	}

// 	for _, node := range lab.Nodes {
// 		// set the configuration IF the node is not yet started
// 		tflog.Info(ctx, fmt.Sprintf("injectConfigs: node state, %+v", node.State))
// 		if node.State == cmlclient.NodeStateDefined {
// 			tflog.Info(ctx, fmt.Sprintf("injectConfigs: defined, %+v", node))

// 			o := cml2Node{}
// 			diags = tfsdk.ValueFrom(ctx, node, types.ObjectType{
// 				AttrTypes: data.nodeAttrs,
// 			}, &o)
// 			if diags.HasError() {
// 				return diags
// 			}

// 			sp := r.matchSpecial(ctx, &diags, special, o)
// 			if diags.HasError() {
// 				tflog.Error(ctx, "ModifyPlan: specials issue")
// 				return diags
// 			}

// 			if sp != nil && !sp.Configuration.IsNull() {
// 				configuration := sp.Configuration.Value
// 				tflog.Info(ctx, fmt.Sprintf("injectConfigs: %s, %s", configuration, node.Configuration))
// 				if configuration == node.Configuration {
// 					continue
// 				}
// 				err := r.client.SetNodeConfig(ctx, lab.ID, node.ID, configuration)
// 				if err != nil {
// 					diags.AddError("set node config failed",
// 						fmt.Sprintf("setting the new node configuration failed: %s", err),
// 					)
// 				}
// 			}
// 		}
// 	}
// 	tflog.Info(ctx, "injectConfigs: done")

// 	return diags
// }

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
	// r.injectConfigs(ctx, lab, data)
	if data.State.Null || data.State.Value == cmlclient.LabStateStarted {
		r.start(ctx, resp.Diagnostics, lab.ID)
	}

	// if unspecified, wait for it converge
	if data.Wait.Null || data.Wait.Value {
		r.converge(ctx, resp.Diagnostics, lab.ID)
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
	// data.Nodes = r.populateNodes(ctx, lab)
	data.Booted = types.Bool{Value: lab.Booted()}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	tflog.Info(ctx, "Create: done")
}

// func (r *LabResource) populateNodes(ctx context.Context, lab *cmlclient.Lab) types.List {
// 	// we want this as a stable sort by node UUID
// 	nodeList := []*cmlclient.Node{}
// 	for _, node := range lab.Nodes {
// 		nodeList = append(nodeList, node)
// 	}
// 	sort.Slice(nodeList, func(i, j int) bool {
// 		return nodeList[i].ID < nodeList[j].ID
// 	})
// 	nodes := types.List{ElemType: types.ObjectType{
// 		AttrTypes: r.nodeAttrs,
// 	}}
// 	for _, node := range nodeList {

// 		newNodeElem := types.Object{}
// 		diags := tfsdk.ValueFrom(
// 			ctx, node, types.ObjectType{
// 				AttrTypes: r.nodeAttrs,
// 			}, &newNodeElem)
// 		diags.Append(diags...)
// 		if diags.HasError() {
// 			panic("uh-oh")
// 		}

// 		nodes.Elems = append(nodes.Elems, newNodeElem)
// 	}
// 	return nodes
// }

func (r *LabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *LabResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Info(ctx, "state:", map[string]interface{}{"data": data})

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
	// data.Nodes = r.populateNodes(ctx, lab)
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

	// special := cml2SpecialMap{}
	// if !configData.Special.IsNull() {
	// 	diags = configData.Special.ElementsAs(ctx, &special, false)
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		tflog.Error(ctx, "Update: that didn't work (1)")
	// 		return
	// 	}
	// 	tflog.Info(ctx, fmt.Sprintf("SPECIAL: %+v\n", special))
	// }

	// nodes := []cml2Node{}
	// diags = stateData.Nodes.ElementsAs(ctx, &nodes, false)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	tflog.Error(ctx, "Update: that didn't work (2)")
	// 	return
	// }

	// for idx, node := range nodes {
	// 	sp := r.matchSpecial(ctx, &diags, special, node)
	// 	if diags.HasError() {
	// 		tflog.Error(ctx, "Update: that didn't work (3)")
	// 		return
	// 	}
	// 	if sp != nil {
	// 		switch node.State.Value {
	// 		case cmlclient.NodeStateDefined:
	// 			if !sp.State.IsNull() {
	// 				if sp.State.Value == cmlclient.NodeStateStarted {

	// 					// modify node configuration state
	// 					ap := tftypes.NewAttributePath().
	// 						WithAttributeName("nodes").
	// 						WithElementKeyInt(idx).
	// 						WithAttributeName("configuration")
	// 					diags = resp.State.SetAttribute(ctx, ap, types.String{
	// 						Unknown: true},
	// 					)
	// 					resp.Diagnostics.Append(diags...)
	// 					if resp.Diagnostics.HasError() {
	// 						tflog.Error(ctx, "ModifyPlan: converged plan has errors")
	// 						return
	// 					}
	// 				}

	// 			}
	// 		}
	// 		if !sp.Configuration.IsNull() {
	// 			node.Configuration.Value = sp.Configuration.Value
	// 		}
	// 		if !sp.State.IsNull() {
	// 			node.State.Value = sp.State.Value
	// 		}
	// 		if !sp.ImageID.IsNull() {
	// 			node.State.Value = sp.ImageID.Value
	// 		}
	// 	}

	// }

	// if !stateData.Special.Equal(planData.Special) {
	// 	lab, err := r.client.GetLab(ctx, planData.Id.Value, false)
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			CML2ErrorLabel,
	// 			fmt.Sprintf("Unable to fetch lab, got error: %s", err),
	// 		)
	// 		return
	// 	}
	// 	r.injectConfigs(ctx, lab, planData)
	// }

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
	// planData.Nodes = r.populateNodes(ctx, lab)
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
