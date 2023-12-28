package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlvalidator"
)

var LabGroupAttrType = map[string]attr.Type{
	"id":         types.StringType,
	"name":       types.StringType,
	"permission": types.StringType,
}

type LabGroupModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Permission types.String `tfsdk:"permission"`
}

func NewLabGroup(ctx context.Context, group *cmlclient.LabGroup, diags *diag.Diagnostics) attr.Value {
	newGroup := LabGroupModel{
		ID:         types.StringValue(group.ID),
		Name:       types.StringValue(group.Name),
		Permission: types.StringValue(group.Permission),
	}
	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newGroup,
			types.ObjectType{AttrTypes: LabGroupAttrType},
			&value,
		)...,
	)
	return value
}

func LabGroup() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Group ID (UUID).",
			Optional:    true,
			Computed:    true,
			// PlanModifiers: []planmodifier.String{
			// 	stringplanmodifier.UseStateForUnknown(),
			// },
		},
		"name": schema.StringAttribute{
			Description: "Descriptive group name.",
			Computed:    true,
			// PlanModifiers: []planmodifier.String{
			// 	stringplanmodifier.UseStateForUnknown(),
			// },
		},
		"permission": schema.StringAttribute{
			MarkdownDescription: "Permission, either `read_only` or `read_write`.",
			Optional:            true,
			Computed:            true,
			// PlanModifiers: []planmodifier.String{
			// 	stringplanmodifier.UseStateForUnknown(),
			// },
			Validators: []validator.String{
				cmlvalidator.GroupPermission{},
			},
		},
	}
}
