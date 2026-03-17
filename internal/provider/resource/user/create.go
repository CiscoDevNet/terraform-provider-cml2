package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		data cmlschema.UserModel
		err  error
	)

	tflog.Info(ctx, "Resource user CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := models.User{}
	user.Username = data.Username.ValueString()
	user.Password = data.Password.ValueString()
	user.Fullname = data.Fullname.ValueString()
	user.Email = data.Email.ValueString()
	user.Description = data.Description.ValueString()
	user.IsAdmin = data.IsAdmin.ValueBool()

	// Groups are reconciled via group membership updates after user creation.
	// The users API can return additional/normalized group IDs and omit the
	// submitted list, which triggers Terraform set correlation errors.
	plannedGroups := userGroupIDsFromSet(ctx, &resp.Diagnostics, data.Groups)
	user.Groups = nil

	if !data.ResourcePool.IsUnknown() && !data.ResourcePool.IsNull() {
		rpRaw := data.ResourcePool.ValueString()
		rpUUID, err := uuid.Parse(rpRaw)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("resource_pool"), "Invalid resource_pool", fmt.Sprintf("resource_pool must be a valid UUID: %s", err))
			return
		}
		if rpUUID.Version() != 4 {
			resp.Diagnostics.AddAttributeError(path.Root("resource_pool"), "Invalid resource_pool", "resource_pool must be a UUIDv4.")
			return
		}
		ptr := models.UUID(rpUUID.String())
		user.ResourcePool = &ptr
	}

	newUser, err := r.cfg.Client().User.Create(ctx, models.UserCreateRequest{UserBase: user.UserBase, Password: user.Password})
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to create user, got error: %s", err),
		)
		return
	}

	r.reconcileGroupMembership(ctx, &resp.Diagnostics, newUser.ID, nil, plannedGroups)
	if resp.Diagnostics.HasError() {
		return
	}
	newUser, err = r.cfg.Client().User.GetByID(ctx, newUser.ID)
	if err != nil {
		resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("Unable to get user, got error: %s", err))
		return
	}

	// need to preserve "write once" values as the read does not return the
	// set password
	newUser.Password = data.Password.ValueString()

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewUser(ctx, &newUser, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.UserAttrType},
			&data,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Resource User CREATE done")
}
