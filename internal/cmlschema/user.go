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

var UserAttrType = map[string]attr.Type{
	"id":            types.StringType,
	"username":      types.StringType,
	"password":      types.StringType,
	"fullname":      types.StringType,
	"email":         types.StringType,
	"description":   types.StringType,
	"is_admin":      types.BoolType,
	"directory_dn":  types.StringType,
	"opt_in":        types.BoolType,
	"resource_pool": types.StringType,
	"groups": types.SetType{
		ElemType: types.StringType,
	},
	"labs": types.SetType{
		ElemType: types.StringType,
	},
}

type UserModel struct {
	ID           types.String `tfsdk:"id"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	Fullname     types.String `tfsdk:"fullname"`
	Email        types.String `tfsdk:"email"`
	Description  types.String `tfsdk:"description"`
	IsAdmin      types.Bool   `tfsdk:"is_admin"`
	DirectoryDN  types.String `tfsdk:"directory_dn"`
	OptIn        types.Bool   `tfsdk:"opt_in"`
	ResourcePool types.String `tfsdk:"resource_pool"`
	Groups       types.Set    `tfsdk:"groups"`
	Labs         types.Set    `tfsdk:"labs"`
}

func NewUser(ctx context.Context, user *cmlclient.User, diags *diag.Diagnostics) attr.Value {
	newUser := UserModel{
		ID:          types.StringValue(user.ID),
		Username:    types.StringValue(user.Username),
		Password:    types.StringValue(user.Password),
		Fullname:    types.StringValue(user.Fullname),
		Email:       types.StringValue(user.Email),
		Description: types.StringValue(user.Description),
		IsAdmin:     types.BoolValue(user.IsAdmin),
		DirectoryDN: types.StringValue(user.DirectoryDN),
		OptIn:       types.BoolValue(user.OptIn),
		Groups:      newStringSet(ctx, user.Groups, diags),
		Labs:        newStringSet(ctx, user.Labs, diags),
	}

	newUser.ResourcePool = types.StringNull()
	if user.ResourcePool != nil {
		newUser.ResourcePool = types.StringValue(*user.ResourcePool)
	}

	var value attr.Value
	diags.Append(
		tfsdk.ValueFrom(
			ctx,
			newUser,
			types.ObjectType{AttrTypes: UserAttrType},
			&value,
		)...,
	)
	return value
}

func User() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "User ID (UUID).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"username": schema.StringAttribute{
			Description: "Login name of the user.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"password": schema.StringAttribute{
			Description: "Password of the user.",
			Required:    true,
			Sensitive:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"fullname": schema.StringAttribute{
			Description: "Full name of the user.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"email": schema.StringAttribute{
			Description: "E-mail address of the user.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"description": schema.StringAttribute{
			Description: "Description of the user.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_admin": schema.BoolAttribute{
			Description: "True if the user has admin rights.",
			Optional:    true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"directory_dn": schema.StringAttribute{
			Description: "Directory DN of the user (when using LDAP).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"groups": schema.SetAttribute{
			Description: "Set of group IDs where the user is member of.",
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"labs": schema.SetAttribute{
			Description: "Set of lab IDs the user owns.",
			Computed:    true,
			ElementType: types.StringType,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"opt_in": schema.BoolAttribute{
			Description: "True if has opted in to sending telemetry data.",
			Computed:    true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"resource_pool": schema.StringAttribute{
			Description: "Resource pool ID, if any.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
