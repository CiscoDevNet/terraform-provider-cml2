package group

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

	group := cmlclient.Group{}
	group.Name = data.Name.ValueString()
	group.Description = data.Description.ValueString()

	memberList := make([]string, 0)
	if !data.Members.IsUnknown() {
		var member types.String
		for _, userID := range data.Members.Elements() {
			tfsdk.ValueAs(ctx, userID, &member)
			el := member.ValueString()
			memberList = append(memberList, el)
		}
	}
	group.Members = memberList

	labList := make([]cmlclient.GroupLab, 0)
	if !data.Labs.IsUnknown() {
		var glModel cmlschema.GroupLabModel
		for _, bb := range data.Labs.Elements() {
			tfsdk.ValueAs(ctx, bb, &glModel)
			el := cmlclient.GroupLab{
				ID:         glModel.ID.ValueString(),
				Permission: glModel.Permission.ValueString(),
			}
			labList = append(labList, el)
		}
	}
	group.Labs = labList

	newGroup, err := r.cfg.Client().GroupCreate(ctx, &group)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to create group, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewGroup(ctx, newGroup, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.GroupAttrType},
			&data,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Group CREATE done")
}
