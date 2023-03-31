package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlvalidator"
)

var GroupLabAttrType = map[string]attr.Type{
	"id":         types.StringType,
	"permission": types.StringType,
}

var GroupAttrType = map[string]attr.Type{
	"id":          types.StringType,
	"description": types.StringType,
	"members": types.SetType{
		ElemType: types.StringType,
	},
	"name": types.StringType,
	"labs": types.SetType{
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
	Members     types.Set    `tfsdk:"members"`
	Name        types.String `tfsdk:"name"`
	Labs        types.Set    `tfsdk:"labs"`
}

func newLabs(ctx context.Context, group *cmlclient.Group, diags *diag.Diagnostics) types.Set {
	if len(group.Labs) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: GroupLabAttrType})
	}
	valueSet := make([]attr.Value, 0)
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
		valueSet = append(valueSet, value)
	}
	newSet, dia := types.SetValue(types.ObjectType{AttrTypes: GroupLabAttrType}, valueSet)
	diags.Append(dia...)
	return newSet
}

func NewGroup(ctx context.Context, group *cmlclient.Group, diags *diag.Diagnostics) attr.Value {

	newGroup := GroupModel{
		ID:          types.StringValue(group.ID),
		Description: types.StringValue(group.Description),
		Name:        types.StringValue(group.Name),
		Members:     newStringSet(ctx, group.Members, diags),
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
			Computed:    true,
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"members": schema.SetAttribute{
			Description: "Set of user IDs who are members of this group.",
			Computed:    true,
			Optional:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Description: "Descriptive group name.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"labs": schema.SetNestedAttribute{
			MarkdownDescription: "Set of labs with their permission which are associated to this group.",
			Computed:            true,
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required:    true,
						Description: "Lab ID (UUID).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"permission": schema.StringAttribute{
						Required:    true,
						Description: "Permission.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							cmlvalidator.GroupPermission{},
						},
					},
				},
			},
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
