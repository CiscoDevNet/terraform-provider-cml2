package group

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rschmied/terraform-provider-cml2/internal/cmlschema"
	"github.com/rschmied/terraform-provider-cml2/internal/common"
)

func (gr *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cmlschema.GroupModel

	tflog.Info(ctx, "Datasource Group READ")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := gr.cfg.Client().GetGroups(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get groups, got error: %s", err),
		)
		return
	}

	groupList := make([]attr.Value, 0)
	for _, group := range groups {
		groupList = append(groupList, cmlschema.NewGroup(
			ctx, group, &resp.Diagnostics),
		)
	}

	// resp.Diagnostics.Append(
	// 	tfsdk.ValueFrom(
	// 		ctx,
	// 		groupList,
	// 		types.ListType{
	// 			ElemType: types.ObjectType{
	// 				AttrTypes: cmlschema.GroupAttrType,
	// 			},
	// 		},
	// 		&data.Group,
	// 	)...,
	// )
	// need an ID
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	data.ID = types.StringValue(uuid.New().String())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Datasource System READ: done")
}
