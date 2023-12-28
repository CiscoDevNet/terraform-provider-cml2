package user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
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

	user := cmlclient.User{}
	user.Username = data.Username.ValueString()
	user.Password = data.Password.ValueString()
	user.Fullname = data.Fullname.ValueString()
	user.Email = data.Email.ValueString()
	user.Description = data.Description.ValueString()
	user.IsAdmin = data.IsAdmin.ValueBool()

	stringList := make([]string, 0)
	if !data.Groups.IsUnknown() {
		var elem types.String
		for _, bb := range data.Groups.Elements() {
			tfsdk.ValueAs(ctx, bb, &elem)
			stringList = append(stringList, elem.ValueString())
		}
	}
	user.Groups = stringList

	user.ResourcePool = nil
	if !data.ResourcePool.IsUnknown() {
		user.ResourcePool = data.ResourcePool.ValueStringPointer()
	}

	newUser, err := r.cfg.Client().UserCreate(ctx, &user)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to create user, got error: %s", err),
		)
		return
	}

	// need to preserve "write once" values as the read does not return the
	// set password
	newUser.Password = data.Password.ValueString()

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewUser(ctx, newUser, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.UserAttrType},
			&data,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Resource User CREATE done")
}
