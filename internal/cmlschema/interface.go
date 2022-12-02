package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

func Interface() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Interface ID (UUID).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"label": schema.StringAttribute{
			Description: "Interface label.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"mac_address": schema.StringAttribute{
			Description: "MAC address.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_connected": schema.BoolAttribute{
			Description: "Is the interface connected to a link?",
			Computed:    true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "interface state (`UP` or `DOWN`).",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"ip4": schema.ListAttribute{
			Description: "IPv4 address list.",
			Computed:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"ip6": schema.ListAttribute{
			Description: "IPv6 address list.",
			Computed:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func NewInterface(ctx context.Context, iface *cmlclient.Interface, diags *diag.Diagnostics) attr.Value {

	ip4List := types.ListNull(types.StringType)
	ip6List := types.ListNull(types.StringType)
	var macAddress types.String

	if iface.Runs() {
		// IPv4 addresses
		list := make([]attr.Value, 0)
		for _, ip := range iface.IP4 {
			list = append(list, types.StringValue(ip))
		}
		ip4List, _ = types.ListValue(types.StringType, list)
		// IPv6 addresses
		list = nil
		for _, ip := range iface.IP6 {
			list = append(list, types.StringValue(ip))
		}
		ip6List, _ = types.ListValue(types.StringType, list)
	}
	if iface.Exists() {
		macAddress = types.StringValue(iface.MACaddress)
	} else {
		macAddress = types.StringNull()
	}

	newIface := InterfaceModel{
		Id:          types.StringValue(iface.ID),
		Label:       types.StringValue(iface.Label),
		State:       types.StringValue(iface.State),
		IsConnected: types.BoolValue(iface.IsConnected),
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
