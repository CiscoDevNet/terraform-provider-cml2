package group

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
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

	group := &cmlclient.Group{
		ID:   data.ID.ValueString(),
		Name: data.Name.ValueString(),
	}

	if !data.Description.IsUnknown() {
		group.Description = data.Description.ValueString()
	}

	members := make([]string, 0)
	if !data.Members.IsUnknown() {
		var user types.String
		for _, elem := range data.Members.Elements() {
			tfsdk.ValueAs(ctx, elem, &user)
			members = append(members, user.ValueString())
		}
	}
	group.Members = members

	labList := make([]cmlclient.GroupLab, 0)
	if !data.Labs.IsUnknown() {
		var model cmlschema.GroupLabModel
		for _, elem := range data.Labs.Elements() {
			tfsdk.ValueAs(ctx, elem, &model)
			lab := cmlclient.GroupLab{
				ID:         model.ID.ValueString(),
				Permission: model.Permission.ValueString(),
			}
			labList = append(labList, lab)
		}
	}
	group.Labs = labList

	updatedGroup, err := r.cfg.Client().GroupUpdate(ctx, group)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to update group, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewGroup(ctx, updatedGroup, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.GroupAttrType},
			&data,
		)...,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Info(ctx, "Resource Group UPDATE done")
}
