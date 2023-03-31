package node

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

func (r NodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var (
		stateData, planData cmlschema.NodeModel
		err                 error
	)

	tflog.Info(ctx, "Resource Node UPDATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := &cmlclient.Node{
		ID:    planData.ID.ValueString(),
		LabID: planData.LabID.ValueString(),
		State: planData.State.ValueString(),
	}

	if !planData.X.IsNull() {
		node.X = int(planData.X.ValueInt64())
	}
	if !planData.Y.IsNull() {
		node.Y = int(planData.Y.ValueInt64())
	}
	if !planData.Label.IsNull() {
		node.Label = planData.Label.ValueString()
	}
	if !planData.Tags.IsNull() {
		var tag types.String
		tags := []string{}
		for _, elem := range planData.Tags.Elements() {
			tfsdk.ValueAs(ctx, elem, &tag)
			tags = append(tags, tag.ValueString())
		}
		node.Tags = tags
	}

	// these can only be changed when the node is DEFINED_ON_CORE
	if stateData.State.ValueString() == cmlclient.NodeStateDefined {
		if !planData.Configuration.IsUnknown() {
			value := planData.Configuration.ValueString()
			node.Configuration = &value
		}
		if !planData.RAM.IsUnknown() {
			node.RAM = int(planData.RAM.ValueInt64())
		}
		if !planData.CPUs.IsUnknown() {
			node.CPUs = int(planData.CPUs.ValueInt64())
		}
		if !planData.CPUlimit.IsUnknown() {
			node.CPUlimit = int(planData.CPUlimit.ValueInt64())
		}
		if !planData.BootDiskSize.IsUnknown() {
			node.BootDiskSize = int(planData.BootDiskSize.ValueInt64())
		}
		if !planData.DataVolume.IsUnknown() {
			node.DataVolume = int(planData.DataVolume.ValueInt64())
		}
		if !planData.ImageDefinition.IsUnknown() {
			node.ImageDefinition = planData.ImageDefinition.ValueString()
		}
	}

	// TODO: groups are missing!

	newNode, err := r.cfg.Client().NodeUpdate(ctx, node)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to update node, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewNode(ctx, newNode, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.NodeAttrType},
			&planData,
		)...,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Node UPDATE done")
}
