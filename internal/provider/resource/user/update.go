package user

import (
	"context"
	"fmt"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
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

	user := &cmlclient.User{
		ID:       data.ID.ValueString(),
		Username: data.Username.ValueString(),
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
		user.ResourcePool = data.ResourcePool.ValueStringPointer()
	}

	groups := make([]string, 0)
	if !data.Groups.IsUnknown() {
		var group types.String
		for _, elem := range data.Groups.Elements() {
			tfsdk.ValueAs(ctx, elem, &group)
			groups = append(groups, group.ValueString())
		}
	}
	user.Groups = groups

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
