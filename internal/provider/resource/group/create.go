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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		data cmlschema.GroupModel
		err  error
	)

	tflog.Info(ctx, "Resource group CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := models.Group{}
	group.Name = data.Name.ValueString()
	group.Description = data.Description.ValueString()

	memberList := make([]models.UUID, 0)
	if !data.Members.IsUnknown() {
		var member types.String
		for _, userID := range data.Members.Elements() {
			tfsdk.ValueAs(ctx, userID, &member)
			memberList = append(memberList, models.UUID(member.ValueString()))
		}
	}
	group.Members = memberList

	// Lab permissions are represented via Associations in the new client model.
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

	newGroup, err := r.cfg.Client().Group.Create(ctx, group)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to create group, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(ctx, cmlschema.NewGroup(ctx, &newGroup, &resp.Diagnostics), types.ObjectType{AttrTypes: cmlschema.GroupAttrType}, &data)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Group CREATE done")
}
