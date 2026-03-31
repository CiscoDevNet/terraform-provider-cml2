package cmlschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlvalidator"
)

// LabGroupAttrType is the attribute type map for LabGroupModel.
var LabGroupAttrType = map[string]attr.Type{
	"id":         types.StringType,
	"permission": types.StringType,
}

// LabGroupModel is the Terraform representation of a lab group permission entry.
type LabGroupModel struct {
	ID         types.String `tfsdk:"id"`
	Permission types.String `tfsdk:"permission"`
}

// NewLabGroup converts a lab group entry into a Terraform value.
func NewLabGroup(ctx context.Context, group *models.LabGroup, diags *diag.Diagnostics) attr.Value { //nolint:staticcheck
	newGroup := LabGroupModel{
		ID:         types.StringValue(string(group.ID)),
		Permission: types.StringValue(string(group.Permission)),
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

// NewLabGroupFromAssociation converts a lab-group association into a Terraform value.
func NewLabGroupFromAssociation(ctx context.Context, assoc *models.Association, diags *diag.Diagnostics) attr.Value {
	newGroup := LabGroupModel{
		ID:         types.StringValue(string(assoc.ID)),
		Permission: types.StringValue(TFGroupPermissionFromAssociationPermissions(assoc.Permissions)),
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

// LabGroup returns the schema for a lab group permission nested object.
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
