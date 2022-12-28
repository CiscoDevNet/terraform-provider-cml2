package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

type LinkModel struct {
	ID         types.String `tfsdk:"id"`
	InterfaceA types.String `tfsdk:"interface_a"`
	InterfaceB types.String `tfsdk:"interface_b"`
	LabID      types.String `tfsdk:"lab_id"`
	Label      types.String `tfsdk:"label"`
	CaptureKey types.String `tfsdk:"link_capture_key"`
	State      types.String `tfsdk:"state"`
	NodeA      types.String `tfsdk:"node_a"`
	NodeB      types.String `tfsdk:"node_b"`
	NodeAslot  types.Int64  `tfsdk:"node_a_slot"`
	NodeBslot  types.Int64  `tfsdk:"node_b_slot"`
}

// with simplified=true
// {
// 	"id": "9d999ee0-1bb7-4b70-a3f2-c043669e9b93",
// 	"node_a": "94c685a4-04f6-467a-bb01-e75a93c3e4b5",
// 	"node_b": "cd5ea0a0-a96a-4c9c-84cd-91251bd34f3e",
// 	"state": "DEFINED_ON_CORE"
// }

// with simplified=false
// {
// 	"id": "9d999ee0-1bb7-4b70-a3f2-c043669e9b93",
// 	"interface_a": "f0cc38c3-f5e9-423a-875a-1c2277c1dbcc",
// 	"interface_b": "f345ea75-fe77-45ff-8097-2c25f4c1a971",
// 	"lab_id": "eb53e679-1ac7-4e47-a120-4ba617c6ffc5",
// 	"label": "unmanaged-switch-0-port0<->server-0-eth0",
// 	"link_capture_key": "7b794958-2b49-42bb-9e05-83f1bf488a06",
// 	"node_a": "94c685a4-04f6-467a-bb01-e75a93c3e4b5",
// 	"node_b": "cd5ea0a0-a96a-4c9c-84cd-91251bd34f3e",
// 	"state": "DEFINED_ON_CORE"
// }

var LinkAttrType = map[string]attr.Type{
	"id":               types.StringType,
	"interface_a":      types.StringType,
	"interface_b":      types.StringType,
	"lab_id":           types.StringType,
	"label":            types.StringType,
	"link_capture_key": types.StringType,
	"state":            types.StringType,
	"node_a":           types.StringType,
	"node_b":           types.StringType,
	"node_a_slot":      types.Int64Type,
	"node_b_slot":      types.Int64Type,
}

func Link() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Link ID (UUID).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"interface_a": schema.StringAttribute{
			Description: "Interface ID containing the node (UUID).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"interface_b": schema.StringAttribute{
			Description: "Interface ID containing the node (UUID).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"lab_id": schema.StringAttribute{
			Description: "Lab ID containing the link (UUID).",
			Required:    true,
		},
		"label": schema.StringAttribute{
			Description: "link label (auto generated).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"link_capture_key": schema.StringAttribute{
			Description: "link capture key (when running).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node_a": schema.StringAttribute{
			Description: "Node (A) attached to link.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node_b": schema.StringAttribute{
			Description: "Node (B) attached to link.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node_a_slot": schema.Int64Attribute{
			Description: "Optional interface slot on node A (src), if not provided use next free.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				int64planmodifier.RequiresReplace(),
			},
		},
		"node_b_slot": schema.Int64Attribute{
			Description: "Optional interface slot on node B (dst), if not provided use next free.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				int64planmodifier.RequiresReplace(),
			},
		},
		"state": schema.StringAttribute{
			Description: "Link state.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func NewLink(ctx context.Context, link *cmlclient.Link, diags *diag.Diagnostics) attr.Value {

	newLink := LinkModel{
		ID:         types.StringValue(link.ID),
		Label:      types.StringValue(link.Label),
		State:      types.StringValue(link.State),
		CaptureKey: types.StringValue(link.PCAPkey),
		LabID:      types.StringValue(link.LabID),
		InterfaceA: types.StringValue(link.SrcID),
		InterfaceB: types.StringValue(link.DstID),
		NodeA:      types.StringValue(link.SrcNode),
		NodeB:      types.StringValue(link.DstNode),
		// -1 is "don't care, use next free"
		NodeAslot: types.Int64Value(-1),
		NodeBslot: types.Int64Value(-1),
	}

	if link.SrcSlot >= 0 {
		newLink.NodeAslot = types.Int64Value(int64(link.SrcSlot))
	}
	if link.DstSlot >= 0 {
		newLink.NodeBslot = types.Int64Value(int64(link.DstSlot))
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newLink,
			types.ObjectType{AttrTypes: LinkAttrType},
			&value,
		)...,
	)
	return value
}
