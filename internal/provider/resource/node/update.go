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
		ID:             planData.ID.ValueString(),
		LabID:          planData.LabID.ValueString(),
		State:          planData.State.ValueString(),
		NodeDefinition: planData.NodeDefinition.ValueString(),
	}

	if !planData.X.IsNull() {
		node.X = int(planData.X.ValueInt64())
	}
	if !planData.Y.IsNull() {
		node.Y = int(planData.Y.ValueInt64())
	}
	if !planData.HideLinks.IsNull() {
		node.HideLinks = bool(planData.HideLinks.ValueBool())
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
		node.Configurations = setNamedConfigsFromData(ctx, resp.Diagnostics, planData)
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

	// work around the fact that updating an external connector will "resolve"
	// the device name (if given, worked in previous versions" with the
	// label... e.g. virbr0 -> NAT, bridge0 -> System Bridge. We want to keep
	// the original value in this case, otherwise we run into inconsistent
	// state!
	if node.NodeDefinition == "external_connector" {
		// this is currently not needed but makes the provider a bit future
		// proof in case external connectors have named configs eventually.
		nnc := cmlschema.NewNamedConfigs(ctx, newNode, &resp.Diagnostics)
		if !planData.Configurations.Equal(nnc) {
			resp.Diagnostics.AddError(
				"External connector configurations",
				fmt.Sprintf("Provide proper external connector configurations, not a device name (deprecated)."),
			)
			return
		}
	}

	// when updating with named configs on, we need to move over the returned
	// named config into the single configuration if it was previously used.
	// tflog.Warn(ctx, "###u", map[string]any{"null": stateData.Configuration.IsNull(), "unknown": stateData.Configuration.IsUnknown(), "len": len(node.Configurations)})
	if !stateData.Configuration.IsUnknown() && len(newNode.Configurations) > 0 {
		newNode.Configuration = &newNode.Configurations[0].Content
		newNode.Configurations = nil
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
