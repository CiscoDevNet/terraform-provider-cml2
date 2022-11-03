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
	Groups      types.List   `tfsdk:"groups"`
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
	"groups": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: GroupAttrType,
		},
	},
}

// Lab returns the schema for the Lab model
func Lab() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Computed:    true,
			Description: "CML lab identifier, a UUID",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"state": {
			Computed:            true,
			Description:         "CML lab state, one of DEFINED_ON_CORE, STARTED or STOPPED",
			MarkdownDescription: "CML lab state, one of `DEFINED_ON_CORE`, `STARTED` or `STOPPED`",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"created": {
			Computed:    true,
			Description: "creation datetime string in ISO8601 format",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"modified": {
			Computed:    true,
			Description: "modification datetime string in ISO8601 format",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"title": {
			Optional:    true,
			Computed:    true,
			Description: "title of the lab",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"owner": {
			Computed:    true,
			Description: "owner of the lab, a UUID4",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"description": {
			Optional:    true,
			Computed:    true,
			Description: "lab description",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"node_count": {
			Computed:    true,
			Description: "number of nodes in the lab",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.Int64Type,
		},
		"link_count": {
			Computed:    true,
			Description: "number of links in the lab",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.Int64Type,
		},
		"notes": {
			Optional:    true,
			Computed:    true,
			Description: "lab notes",
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Type: types.StringType,
		},
		"groups": {
			Optional:    true,
			Computed:    true,
			Description: "lab notes",
			Attributes: tfsdk.ListNestedAttributes(
				map[string]tfsdk.Attribute{
					"id": {
						Description: "group ID (UUID)",
						Type:        types.StringType,
						Computed:    true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							resource.UseStateForUnknown(),
						},
					},
					"permission": {
						Description:         "permission, either read_only or read_write",
						MarkdownDescription: "permission, either `read_only` or `read_write`",
						Type:                types.StringType,
						Computed:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							resource.UseStateForUnknown(),
						},
					},
				}),
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
	}
}

// NewLab creates a TF value from a CML2 lab object from the gocmlclient
func NewLab(ctx context.Context, lab *cmlclient.Lab, diags *diag.Diagnostics) attr.Value {

	valueList := make([]attr.Value, 0)
	for _, group := range lab.Groups {
		value := NewGroup(ctx, group, diags)
		valueList = append(valueList, value)
	}
	groups, _ := types.ListValue(
		types.ObjectType{AttrTypes: GroupAttrType},
		valueList,
	)

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
