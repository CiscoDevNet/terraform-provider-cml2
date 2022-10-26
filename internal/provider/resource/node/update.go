package node

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/terraform-provider-cml2/internal/schema"
)

func (r NodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var (
		stateData, planData *schema.NodeModel
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
		ID:    planData.ID.Value,
		LabID: planData.LabID.Value,
	}

	if !planData.X.IsNull() {
		node.X = int(planData.X.Value)
	}
	if !planData.Y.IsNull() {
		node.Y = int(planData.Y.Value)
	}
	if !planData.Label.IsNull() {
		node.Label = planData.Label.Value
	}
	if !planData.Tags.IsNull() {
		tags := []string{}
		tag := types.String{}
		for _, elem := range planData.Tags.Elems {
			// Ignore error and diagnostics for the simple conversion here
			// Can't use elem.String() here as that has the value in quotes!
			tfsdk.ValueAs(ctx, elem, &tag)
			tags = append(tags, tag.Value)
		}
		node.Tags = tags
	}

	// these can only be changed when the node is DEFINED_ON_CORE
	if stateData.State.Value == cmlclient.LabStateDefined {
		if !planData.Configuration.IsNull() {
			node.Configuration = planData.Configuration.Value
		}
		if !planData.RAM.IsNull() {
			node.RAM = int(planData.RAM.Value)
		}
		if !planData.CPUs.IsNull() {
			node.CPUs = int(planData.CPUs.Value)
		}
		if !planData.CPUlimit.IsNull() {
			node.CPUlimit = int(planData.CPUlimit.Value)
		}
		if !planData.BootDiskSize.IsNull() {
			node.BootDiskSize = int(planData.BootDiskSize.Value)
		}
		if !planData.DataVolume.IsNull() {
			node.DataVolume = int(planData.DataVolume.Value)
		}
		if !planData.ImageDefinition.IsNull() {
			node.ImageDefinition = planData.ImageDefinition.Value
		}
	}

	// TODO: groups are missing!

	newNode, err := r.cfg.Client().NodeUpdate(ctx, node)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to update node, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			schema.NewNode(ctx, newNode, &resp.Diagnostics),
			types.ObjectType{AttrTypes: schema.NodeAttrType},
			&planData,
		)...,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)

	tflog.Info(ctx, "Resource Node UPDATE: done")
}
