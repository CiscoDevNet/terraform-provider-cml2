package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

type InterfaceModel struct {
	Id          types.String `tfsdk:"id"`
	Label       types.String `tfsdk:"label"`
	State       types.String `tfsdk:"state"`
	MACaddress  types.String `tfsdk:"mac_address"`
	IsConnected types.Bool   `tfsdk:"is_connected"`
	IP4         types.List   `tfsdk:"ip4"`
	IP6         types.List   `tfsdk:"ip6"`
}

var InterfaceAttrType = map[string]attr.Type{
	"id":           types.StringType,
	"label":        types.StringType,
	"state":        types.StringType,
	"mac_address":  types.StringType,
	"is_connected": types.BoolType,
	"ip4":          types.ListType{ElemType: types.StringType},
	"ip6":          types.ListType{ElemType: types.StringType},
}

func Interface() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Description: "interface ID (UUID)",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"label": {
			Description: "label",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"mac_address": {
			Description: "MAC address",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"is_connected": {
			Description: "connection status",
			Type:        types.BoolType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"state": {
			Description: "interface state (UP / DOWN)",
			Type:        types.StringType,
			Computed:    true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ip4": {
			Description: "IPv4 address list",
			Computed:    true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"ip6": {
			Description: "IPv6 address list",
			Computed:    true,
			Type: types.ListType{
				ElemType: types.StringType,
			},
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
	}
}

func NewInterface(ctx context.Context, iface *cmlclient.Interface, diags *diag.Diagnostics) attr.Value {

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
		macAddress = types.String{Value: iface.MACaddress}
	} else {
		macAddress = types.String{Null: true, Unknown: true}
	}

	newIface := InterfaceModel{
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
			types.ObjectType{AttrTypes: InterfaceAttrType},
			&value,
		)...,
	)
	return value
}
