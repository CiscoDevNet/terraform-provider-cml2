package group

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		data, state cmlschema.GroupModel
		err         error
	)

	tflog.Info(ctx, "Resource Group UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := models.Group{
		ID:   models.UUID(data.ID.ValueString()),
		Name: data.Name.ValueString(),
	}

	if !data.Description.IsUnknown() {
		group.Description = data.Description.ValueString()
	}

	members := make([]models.UUID, 0)
	if !data.Members.IsUnknown() {
		var user types.String
		for _, elem := range data.Members.Elements() {
			tfsdk.ValueAs(ctx, elem, &user)
			members = append(members, models.UUID(user.ValueString()))
		}
	}
	group.Members = members

	assocs := make([]models.Association, 0)
	if !data.Labs.IsUnknown() && !data.Labs.IsNull() {
		var lab cmlschema.GroupLabModel
		for _, elem := range data.Labs.Elements() {
			resp.Diagnostics.Append(tfsdk.ValueAs(ctx, elem, &lab)...)
			if resp.Diagnostics.HasError() {
				return
			}
			assocs = append(assocs, models.Association{
				ID:          models.UUID(lab.ID.ValueString()),
				Permissions: cmlschema.AssociationPermissionsFromTFGroupPermission(lab.Permission.ValueString()),
			})
		}
	}
	group.Associations = assocs

	updatedGroup, err := r.cfg.Client().Group.Update(ctx, group)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to update group, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(ctx, cmlschema.NewGroup(ctx, &updatedGroup, &resp.Diagnostics), types.ObjectType{AttrTypes: cmlschema.GroupAttrType}, &data)...,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Resource Group UPDATE done")
}
