package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

var GroupLabAttrType = map[string]attr.Type{
	"id":         types.StringType,
	"permission": types.StringType,
}

var GroupAttrType = map[string]attr.Type{
	"id":          types.StringType,
	"description": types.StringType,
	"members": types.ListType{
		ElemType: types.StringType,
	},
	"name": types.StringType,
	"labs": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: GroupLabAttrType,
		},
	},
}

type GroupLabModel struct {
	ID         types.String `tfsdk:"id"`
	Permission types.String `tfsdk:"permission"`
}

type GroupModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Members     types.List   `tfsdk:"members"`
	Name        types.String `tfsdk:"name"`
	Labs        types.List   `tfsdk:"labs"`
}

func newMembers(ctx context.Context, members []string, diags *diag.Diagnostics) types.List {
	if len(members) == 0 {
		return types.ListNull(types.StringType)
	}
	valueList := make([]attr.Value, 0)
	for _, member := range members {
		valueList = append(valueList, types.StringValue(member))
	}
	memberList, dia := types.ListValue(types.StringType, valueList)
	diags.Append(dia...)
	return memberList
}

func newLabs(ctx context.Context, group *cmlclient.Group, diags *diag.Diagnostics) types.List {
	if len(group.Labs) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: GroupLabAttrType})
	}
	valueList := make([]attr.Value, 0)
	for _, lab := range group.Labs {
		var value attr.Value

		newLab := GroupLabModel{
			ID:         types.StringValue(lab.ID),
			Permission: types.StringValue(lab.Permission),
		}

		diags.Append(tfsdk.ValueFrom(
			ctx,
			newLab,
			types.ObjectType{AttrTypes: GroupLabAttrType},
			&value,
		)...)
		valueList = append(valueList, value)
	}
	labList, dia := types.ListValue(types.ObjectType{AttrTypes: GroupLabAttrType}, valueList)
	diags.Append(dia...)
	return labList
}

func NewGroup(ctx context.Context, group *cmlclient.Group, diags *diag.Diagnostics) attr.Value {

	newGroup := GroupModel{
		ID:          types.StringValue(group.ID),
		Description: types.StringValue(group.Description),
		Name:        types.StringValue(group.Name),
		Members:     newMembers(ctx, group.Members, diags),
		Labs:        newLabs(ctx, group, diags),
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newGroup,
			types.ObjectType{AttrTypes: GroupAttrType},
			&value,
		)...,
	)
	return value
}

func Group() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Group ID (UUID).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"description": schema.StringAttribute{
			Description: "Description of the group.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"members": schema.ListAttribute{
			Description: "List of user IDs who are members of this group.",
			Optional:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Description: "Descriptive group name.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"labs": schema.ListNestedAttribute{
			MarkdownDescription: "List of labs with their permission which are associated to this group.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Lab ID (UUID).",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"permission": schema.StringAttribute{
						Description: "Permission.",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
