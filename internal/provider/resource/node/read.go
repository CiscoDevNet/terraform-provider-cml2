package node

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

func (r *NodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cmlschema.NodeModel

	tflog.Info(ctx, "Resource Node READ")

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := &cmlclient.Node{
		LabID: data.LabID.ValueString(),
		ID:    data.ID.ValueString(),
	}
	node, err := r.cfg.Client().NodeGet(ctx, node)
	if err != nil {
		resp.Diagnostics.AddError(
			common.ErrorLabel,
			fmt.Sprintf("Unable to get node, got error: %s", err),
		)
		return
	}

	// tflog.Warn(ctx, "###1", map[string]any{"null": data.Configuration.IsNull(), "unknown": data.Configuration.IsUnknown(), "len": len(node.Configurations)})
	if !data.Configuration.IsNull() && len(node.Configurations) > 0 {
		node.Configuration = &node.Configurations[0].Content
		node.Configurations = nil
	}

	resp.Diagnostics.Append(
		tfsdk.ValueFrom(
			ctx,
			cmlschema.NewNode(ctx, node, &resp.Diagnostics),
			types.ObjectType{AttrTypes: cmlschema.NodeAttrType},
			&data,
		)...,
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Resource Node READ done")
}
