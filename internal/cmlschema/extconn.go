package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

// ExtConnModel is the TF representation of a CML2 External Connector (no
// operational data at the moment)
type ExtConnModel struct {
	ID         types.String `tfsdk:"id"`
	DeviceName types.String `tfsdk:"device_name"`
	Label      types.String `tfsdk:"label"`
	Protected  types.Bool   `tfsdk:"protected"`
	Snooped    types.Bool   `tfsdk:"snooped"`
	Tags       types.Set    `tfsdk:"tags"`
}

// ExtConnAttrType has the attribute types of a CML2 ExtConnModel
var ExtConnAttrType = map[string]attr.Type{
	"id":          types.StringType,
	"device_name": types.StringType,
	"label":       types.StringType,
	"protected":   types.BoolType,
	"snooped":     types.BoolType,
	"tags": types.SetType{
		ElemType: types.StringType,
	},
}

// ExtConn returns the schema for the ExtConn model
func ExtConn() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "External connector identifier, a UUID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"device_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "the actual (Linux network) device name of the external connector.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"label": schema.StringAttribute{
			Computed:    true,
			Description: "The label of the external connector, like \"NAT\" or \"System Bridge\".",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"protected": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the connector is protected, e.g. BPDUs are filtered or not.",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"snooped": schema.BoolAttribute{
			Computed:    true,
			Description: "True if the IP address snooper listens on this connector.",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"tags": schema.SetAttribute{
			Description: "The external connector tag set.",
			Computed:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// NewExtConn creates a TF value from a CML2 extconn object from the gocmlclient
func NewExtConn(ctx context.Context, extconn *cmlclient.ExtConn, diags *diag.Diagnostics) attr.Value {
	valueSet := make([]attr.Value, 0)
	for _, tag := range extconn.Tags {
		valueSet = append(valueSet, types.StringValue(tag))
	}
	tags, diag := types.SetValue(
		types.StringType,
		valueSet,
	)
	diags.Append(diag...)

	newConnector := ExtConnModel{
		ID:         types.StringValue(extconn.ID),
		DeviceName: types.StringValue(extconn.DeviceName),
		Label:      types.StringValue(extconn.Label),
		Protected:  types.BoolValue(extconn.Protected),
		Snooped:    types.BoolValue(extconn.Snooped),
		Tags:       tags,
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newConnector,
			types.ObjectType{AttrTypes: ExtConnAttrType},
			&value,
		)...,
	)
	return value
}
