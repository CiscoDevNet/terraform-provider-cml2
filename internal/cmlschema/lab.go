package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

// LabModel is the TF representation of a CML2 lab
type LabModel struct {
	ID          types.String `tfsdk:"id"`
	State       types.String `tfsdk:"state"`
	Created     types.String `tfsdk:"created"`
	Modified    types.String `tfsdk:"modified"`
	Title       types.String `tfsdk:"title"`
	Owner       types.String `tfsdk:"owner"`
	Description types.String `tfsdk:"description"`
	NodeCount   types.Int64  `tfsdk:"node_count"`
	LinkCount   types.Int64  `tfsdk:"link_count"`
	Notes       types.String `tfsdk:"notes"`
	Groups      types.Set    `tfsdk:"groups"`
}

// LabAttrType has the attribute types of a CML2 LabModel
var LabAttrType = map[string]attr.Type{
	"id":          types.StringType,
	"state":       types.StringType,
	"created":     types.StringType,
	"modified":    types.StringType,
	"title":       types.StringType,
	"owner":       types.StringType,
	"description": types.StringType,
	"node_count":  types.Int64Type,
	"link_count":  types.Int64Type,
	"notes":       types.StringType,
	"groups": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: LabGroupAttrType,
		},
	},
}

// Lab returns the schema for the Lab model
func Lab() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Lab identifier, a UUID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"state": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Lab state, one of `DEFINED_ON_CORE`, `STARTED` or `STOPPED`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"created": schema.StringAttribute{
			Computed:    true,
			Description: "Creation date/time string in ISO8601 format.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"modified": schema.StringAttribute{
			Computed:    true,
			Description: "Modification date/time string in ISO8601 format.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"title": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Title of the lab.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"owner": schema.StringAttribute{
			Computed:    true,
			Description: "Owner of the lab, a UUID4.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"description": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Lab description.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node_count": schema.Int64Attribute{
			Computed:    true,
			Description: "Number of nodes in the lab.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"link_count": schema.Int64Attribute{
			Computed:    true,
			Description: "Number of links in the lab.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"notes": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Lab notes.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"groups": schema.SetNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Groups assigned to the lab.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: LabGroup(),
			},
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// NewLab creates a TF value from a CML2 lab object from the gocmlclient
func NewLab(ctx context.Context, lab *cmlclient.Lab, diags *diag.Diagnostics) attr.Value {

	valueSet := make([]attr.Value, 0)
	for _, group := range lab.Groups {
		value := NewLabGroup(ctx, group, diags)
		valueSet = append(valueSet, value)
	}
	groups, diag := types.SetValue(
		types.ObjectType{AttrTypes: LabGroupAttrType},
		valueSet,
	)
	diags.Append(diag...)

	newLab := LabModel{
		ID:          types.StringValue(lab.ID),
		State:       types.StringValue(lab.State),
		Created:     types.StringValue(lab.Created),
		Modified:    types.StringValue(lab.Modified),
		Title:       types.StringValue(lab.Title),
		Owner:       types.StringValue(lab.Owner.ID),
		Description: types.StringValue(lab.Description),
		NodeCount:   types.Int64Value(int64(lab.NodeCount)),
		LinkCount:   types.Int64Value(int64(lab.LinkCount)),
		Notes:       types.StringValue(lab.Notes),
		Groups:      groups,
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newLab,
			types.ObjectType{AttrTypes: LabAttrType},
			&value,
		)...,
	)
	return value
}
