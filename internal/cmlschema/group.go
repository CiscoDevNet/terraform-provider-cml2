package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
)

var GroupAttrType = map[string]attr.Type{
	"id":         types.StringType,
	"permission": types.StringType,
}

type GroupModel struct {
	ID         types.String `tfsdk:"id"`
	Permission types.String `tfsdk:"permission"`
}

func NewGroup(ctx context.Context, group *cmlclient.Group, diags *diag.Diagnostics) attr.Value {

	newGroup := GroupModel{
		ID:         types.StringValue(group.ID),
		Permission: types.StringValue(group.Permission),
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
		"permission": schema.StringAttribute{
			MarkdownDescription: "Permission, either `read_only` or `read_write`.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
