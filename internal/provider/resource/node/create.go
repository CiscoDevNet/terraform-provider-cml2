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

func (r *NodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		data cmlschema.NodeModel
		err  error
	)

	tflog.Info(ctx, "Resource Node CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := cmlclient.Node{}

	node.LabID = data.LabID.ValueString()

	if !data.Label.IsUnknown() {
		node.Label = data.Label.ValueString()
	}
	if !data.NodeDefinition.IsUnknown() {
		node.NodeDefinition = data.NodeDefinition.ValueString()
	}

	// We always need to create a tag list as the API always returns a list of
	// tags even if none are set... e.g. no tags --> [] (instead of null).
	tags := []string{}
	for _, tag := range data.Tags.Elements() {
		tags = append(tags, tag.(types.String).ValueString())
	}
	node.Tags = tags

	if !data.Configuration.IsUnknown() {
		value := data.Configuration.ValueString()
		node.Configuration = &value
	}
	if !data.X.IsUnknown() {
		node.X = int(data.X.ValueInt64())
	}
	if !data.Y.IsUnknown() {
		node.Y = int(data.Y.ValueInt64())
	}
	if !data.HideLinks.IsUnknown() {
		node.HideLinks = bool(data.HideLinks.ValueBool())
	}
	if !data.RAM.IsUnknown() {
		node.RAM = int(data.RAM.ValueInt64())
	}
	if !data.CPUs.IsUnknown() {
		node.CPUs = int(data.CPUs.ValueInt64())
	}
	if !data.CPUlimit.IsUnknown() {
		node.CPUlimit = int(data.CPUlimit.ValueInt64())
	}
	if !data.BootDiskSize.IsUnknown() {
		node.BootDiskSize = int(data.BootDiskSize.ValueInt64())
	}
	if !data.DataVolume.IsUnknown() {
		node.DataVolume = int(data.DataVolume.ValueInt64())
	}
	if !data.ImageDefinition.IsUnknown() {
		node.ImageDefinition = data.ImageDefinition.ValueString()
	}

	// can't set a configuration for an unmanaged switch
	if node.NodeDefinition == "unmanaged_switch" && !data.Configuration.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unmanaged switch configuration",
			"Can't provide UMS configuration",
		)
		return
	}

	newNode, err := r.cfg.Client().NodeCreate(ctx, &node)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to create node, got error: %s", err),
		)
		return
	}

	// work around the fact that creating an external connector will "resolve"
	// the device name (if given, worked in previous versions" with the
	// label... e.g. virbr0 -> NAT, bridge0 -> System Bridge. We return an
	// error in this case, otherwise we'd run into inconsistent state!
	if node.NodeDefinition == "external_connector" && node.Configuration != nil && *newNode.Configuration != *node.Configuration {
		resp.Diagnostics.AddError(
			"External connector configuration",
			fmt.Sprintf("provide proper external connector configuration, not a device name (deprecated). API returned %q, configured was %q!", *newNode.Configuration, *node.Configuration),
		)
		return
		// newNode.Configuration = node.Configuration
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewNode(ctx, newNode, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.NodeAttrType},
			&data,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Node CREATE done")
}
