package group

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

	if !data.Members.IsUnknown() {
		var memberList []string
		var member types.String
		for _, userID := range data.Members.Elements() {
			tfsdk.ValueAs(ctx, userID, &member)
			el := member.ValueString()
			memberList = append(memberList, el)
		}
		group.Members = memberList
	}

	if !data.Labs.IsUnknown() {
		var labList []cmlclient.GroupLab
		var glModel cmlschema.GroupLabModel
		for _, bb := range data.Labs.Elements() {
			tfsdk.ValueAs(ctx, bb, &glModel)
			el := cmlclient.GroupLab{
				ID:         glModel.ID.ValueString(),
				Permission: glModel.Permission.ValueString(),
			}
			labList = append(labList, el)
		}
		group.Labs = labList
	}

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
