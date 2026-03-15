package user

import (
	"context"
	"fmt"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		data, state cmlschema.UserModel
		err         error
	)

	tflog.Info(ctx, "Resource User UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := &models.User{
		ID: models.UUID(data.ID.ValueString()),
		UserBase: models.UserBase{
			Username: data.Username.ValueString(),
		},
		// passwords can't be changed by just setting the new password
		Password: "",
	}

	if !data.Fullname.IsUnknown() {
		user.Fullname = data.Fullname.ValueString()
	}

	if !data.Email.IsUnknown() {
		user.Email = data.Email.ValueString()
	}

	if !data.Description.IsUnknown() {
		user.Description = data.Description.ValueString()
	}

	if !data.IsAdmin.IsUnknown() {
		user.IsAdmin = data.IsAdmin.ValueBool()
	}

	if !data.ResourcePool.IsUnknown() {
		if data.ResourcePool.IsNull() {
			user.ResourcePool = nil
		} else {
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
	}

	plannedGroups := userGroupIDsFromSet(ctx, &resp.Diagnostics, data.Groups)
	user.Groups = nil

	// can't update password
	user.Password = ""

	updatedUser, err := r.cfg.Client().UserUpdate(ctx, user)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to update user, got error: %s", err),
		)
		return
	}

	if !data.Groups.IsUnknown() {
		r.reconcileGroupMembership(ctx, &resp.Diagnostics, updatedUser.ID, stateGroupIDsFromSet(ctx, &resp.Diagnostics, state.Groups), plannedGroups)
		if resp.Diagnostics.HasError() {
			return
		}
		updatedUser, err = r.cfg.Client().UserGet(ctx, string(updatedUser.ID))
		if err != nil {
			resp.Diagnostics.AddError(common.ErrorLabel, fmt.Sprintf("Unable to get user, got error: %s", err))
			return
		}
	}

	// if state.Password.ValueString() != data.Password.ValueString() {
	// 	fmt.Println("BLA")
	// }
	// need to preserve "write once" values
	updatedUser.Password = data.Password.ValueString()

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewUser(ctx, updatedUser, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.UserAttrType},
			&data,
		)...,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Resource User UPDATE done")
}
