package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
		ID:         types.String{Value: group.ID},
		Permission: types.String{Value: group.Permission},
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
