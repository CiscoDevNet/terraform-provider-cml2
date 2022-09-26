package provider

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

type LabResourceModel struct {
	Topology types.String `tfsdk:"topology"`
	Wait     types.Bool   `tfsdk:"wait"`
	Id       types.String `tfsdk:"id"`
	State    types.String `tfsdk:"state"`
	Booted   types.Bool   `tfsdk:"booted"`
	Nodes    types.Map    `tfsdk:"nodes"`
	Configs  types.Map    `tfsdk:"configs"`
	Staging  types.Object `tfsdk:"staging"`
	Timeouts types.Object `tfsdk:"timeouts"`
}

type ResourceTimeouts struct {
	Create types.String `tfsdk:"create"`
	Update types.String `tfsdk:"update"`
	Delete types.String `tfsdk:"delete"`
}

type ResourceStaging struct {
	Stages    types.List   `tfsdk:"stages"`
	Unmatched types.String `tfsdk:"unmatched"`
}

type NodeResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Label          types.String `tfsdk:"label"`
	State          types.String `tfsdk:"state"`
	NodeDefinition types.String `tfsdk:"nodedefinition"`
	Configuration  types.String `tfsdk:"configuration"`
	Interfaces     types.List   `tfsdk:"interfaces"`
	Tags           types.List   `tfsdk:"tags"`
}

type InterfaceResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Label       types.String `tfsdk:"label"`
	State       types.String `tfsdk:"state"`
	MACaddress  types.String `tfsdk:"mac_address"`
	IsConnected types.Bool   `tfsdk:"is_connected"`
	IP4         types.List   `tfsdk:"ip4"`
	IP6         types.List   `tfsdk:"ip6"`
}

var interfaceAttrType = map[string]attr.Type{
	"id":           types.StringType,
	"label":        types.StringType,
	"state":        types.StringType,
	"mac_address":  types.StringType,
	"is_connected": types.BoolType,
	"ip4":          types.ListType{ElemType: types.StringType},
	"ip6":          types.ListType{ElemType: types.StringType},
}

var nodeAttrType = map[string]attr.Type{
	"id":             types.StringType,
	"label":          types.StringType,
	"state":          types.StringType,
	"nodedefinition": types.StringType,
	"configuration":  types.StringType,
	"interfaces": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: interfaceAttrType,
		},
	},
	"tags": types.ListType{ElemType: types.StringType},
}

func newNode(ctx context.Context, node *cmlclient.Node, diags *diag.Diagnostics) attr.Value {

	// we want this as a stable sort by interface UUID
	ilist := []*cmlclient.Interface{}
	for _, iface := range node.Interfaces {
		ilist = append(ilist, iface)
	}
	sort.Slice(ilist, func(i, j int) bool {
		return ilist[i].ID < ilist[j].ID
	})

	ifaces := types.List{ElemType: types.ObjectType{
		AttrTypes: interfaceAttrType,
	}}
	for _, iface := range ilist {
		value := newInterface(ctx, iface, diags)
		ifaces.Elems = append(ifaces.Elems, value)
	}

	tags := types.List{ElemType: types.StringType}
	for _, tag := range node.Tags {
		tags.Elems = append(tags.Elems, types.String{Value: tag})
	}

	newNode := NodeResourceModel{
		Id:             types.String{Value: node.ID},
		Label:          types.String{Value: node.Label},
		State:          types.String{Value: node.State},
		NodeDefinition: types.String{Value: node.NodeDefinition},
		Configuration:  types.String{Value: node.Configuration},
		Interfaces:     ifaces,
		Tags:           tags,
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newNode,
			types.ObjectType{AttrTypes: nodeAttrType},
			&value,
		)...,
	)
	return value
}

func newInterface(ctx context.Context, iface *cmlclient.Interface, diags *diag.Diagnostics) attr.Value {

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
		macAddress.Unknown = false
	} else {
		macAddress.Unknown = true
		macAddress.Null = true
	}

	newIface := InterfaceResourceModel{
		Id:          types.String{Value: iface.ID},
		Label:       types.String{Value: iface.Label},
		State:       types.String{Value: iface.State},
		IsConnected: types.Bool{Value: iface.IsConnected},
		MACaddress:  macAddress,
		IP4:         ip4List,
		IP6:         ip6List,
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newIface,
			types.ObjectType{AttrTypes: interfaceAttrType},
			&value,
		)...,
	)
	return value
}
