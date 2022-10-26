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

func (r *NodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var (
		data *schema.NodeModel
		err  error
	)

	tflog.Info(ctx, "Resource Node CREATE")

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// label
	// nodedefiniton
	// imagedefinition
	// tags
	// configuration
	// x, y
	// cpus, cpu_limit
	// boot_disk_size, data_volume

	node := cmlclient.Node{}

	node.LabID = data.LabID.Value

	if !data.Label.IsNull() {
		node.Label = data.Label.Value
	}
	if !data.NodeDefinition.IsNull() {
		node.NodeDefinition = data.NodeDefinition.Value
	}
	if !data.ImageDefinition.IsNull() {
		node.ImageDefinition = data.ImageDefinition.Value
	}
	if !data.Tags.IsNull() {
		tags := []string{}
		for _, tag := range data.Tags.Elems {
			tags = append(tags, tag.(types.String).Value)
		}
		node.Tags = tags
	}
	if !data.Configuration.IsNull() {
		node.Configuration = data.Configuration.Value
	}
	if !data.X.IsNull() {
		node.X = int(data.X.Value)
	}
	if !data.Y.IsNull() {
		node.Y = int(data.Y.Value)
	}
	if !data.CPUs.IsNull() {
		node.CPUs = int(data.CPUs.Value)
	}
	if !data.CPUlimit.IsNull() {
		node.CPUlimit = int(data.CPUlimit.Value)
	}
	if !data.BootDiskSize.IsNull() {
		node.BootDiskSize = int(data.BootDiskSize.Value)
	}
	if !data.DataVolume.IsNull() {
		node.DataVolume = int(data.DataVolume.Value)
	}
	if !data.RAM.IsNull() {
		node.RAM = int(data.RAM.Value)
	}

	newNode, err := r.cfg.Client().NodeCreate(ctx, &node)
	if err != nil {
		resp.Diagnostics.AddError(
			CML2ErrorLabel,
			fmt.Sprintf("Unable to create node, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			schema.NewNode(ctx, newNode, &resp.Diagnostics),
			types.ObjectType{AttrTypes: schema.NodeAttrType},
			&data,
		)...,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Node CREATE: done")
}
